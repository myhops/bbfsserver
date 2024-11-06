package main

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/myhops/bbfs"
	"github.com/myhops/bbfsserver/handlers/cache"
	"github.com/myhops/bbfsserver/handlers/settable"
	"github.com/myhops/bbfsserver/resources"
	"github.com/myhops/bbfsserver/server"
)

type resetServer struct {
	http.Server
	handler *settable.Settable

	// Keep for rebuilds
	logger *slog.Logger
	opts   *options

	// last tag
	lastTag string
}

func (s *resetServer) buildHandler(logger *slog.Logger, opts *options) (http.Handler, error) {
	cfg := &bbfs.Config{
		Host:           opts.host,
		ProjectKey:     opts.projectKey,
		RepositorySlug: opts.repositorySlug,
		AccessKey:      opts.accessKey,
	}

	allFS := bbfs.NewFS(cfg)

	var versions []*server.Version
	versions = getDryRunVersions(cfg, logger)
	if opts.dryRun != "true" {
		v, err := getVersions(cfg, logger)
		if err != nil {
			return nil, fmt.Errorf("error getting tags: %w", err)
		}
		versions = v
	}

	s.lastTag = getLatestTag(opts, logger)
	tags, err := getTags(cfg, logger)
	if err != nil {
		return nil, err
	}
	lastTag := ""
	if len(tags) > 0 {
		lastTag = tags[0]
	}

	getinfo := getIndexPageInfo(
		opts.repoURL,
		"OLO KOR Build Reports",
		opts.projectKey,
		opts.repositorySlug,
		tags,
	)

	s.lastTag = lastTag

	webFS, err := fs.Sub(resources.StaticHtmlFS, "web")
	if err != nil {
		return nil, fmt.Errorf("error creating web sub fs: %w", err)
	}

	vfsh := server.New(
		logger,
		allFS,
		versions,
		webFS,
		resources.IndexHtmlTemplate,
		getinfo,
		opts.changePollingInterval,
		cache.Middleware(10_000),
	)
	return vfsh, nil
}

func getLatestTag(opts *options, logger *slog.Logger) string {
	cfg := &bbfs.Config{
		Host:           opts.host,
		ProjectKey:     opts.projectKey,
		RepositorySlug: opts.repositorySlug,
		AccessKey:      opts.accessKey,
	}
	tags, err := getTags(cfg, logger)
	if err != nil {
		return ""
	}
	if len(tags) == 0 {
		return ""
	}
	return tags[0]
}

func newServer(ctx context.Context, logger *slog.Logger, opts *options) (*resetServer, error) {
	// baseContext for the http server
	baseContext := func(_ net.Listener) context.Context {
		return ctx
	}

	srv := &resetServer{
		Server: http.Server{
			Addr:              opts.listenAddress,
			ReadHeaderTimeout: 10 * time.Second,
			BaseContext:       baseContext,
		},
		logger: logger,
		opts:   opts,
	}
	h, err := srv.buildHandler(logger, opts)
	if err != nil {
		return nil, fmt.Errorf("build hander failed: %s", err.Error())
	}

	h = LogRequestMiddleware(h.ServeHTTP, logger)

	sh := settable.New(h)

	// create the server
	srv.handler = sh
	return srv, nil
}

func (s *resetServer) rebuild() error {
	s.logger.Info("rebuilding server")
	h, err := s.buildHandler(s.logger, s.opts)
	if err != nil {
		s.logger.Error("building handler failed", slog.String("error", err.Error()))
		return fmt.Errorf("build hander failed: %s", err.Error())
	}

	h = LogRequestMiddleware(h.ServeHTTP, s.logger)

	// Set the handler
	s.handler.Set(h)
	s.logger.Info("set new handler")

	return nil
}
