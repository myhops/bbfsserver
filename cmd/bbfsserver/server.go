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

// type serverConfig struct {
// 	Logger          *slog.Logger
// 	AllFS           fs.FS
// 	Versions        []*server.Version
// 	WebFS           fs.FS
// 	IndexTemplate   string
// 	GetInfo         func() (*server.IndexPageInfo, error)
// 	TimeToLive      time.Duration
// 	CacheMiddleware func(next http.Handler) http.Handler
// }

type resetServer struct {
	http.Server
	handler *settable.Settable

	// Keep for rebuilds
	logger *slog.Logger
	opts   *options
}

func buildHandler(logger *slog.Logger, opts *options) (http.Handler, error) {
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

	getinfo := getIndexPageInfo(
		opts.repoURL,
		"OLO KOR Build Reports",
		opts.projectKey,
		opts.repositorySlug,
		getTagsNil(cfg, logger),
	)

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
		opts.tagsPollInterval,
		cache.Middleware(10_000),
	)
	return vfsh, nil
}

func newServer(ctx context.Context, logger *slog.Logger, opts *options) (*resetServer, error) {
	h, err := buildHandler(logger, opts)
	if err != nil {
		return nil, fmt.Errorf("build hander failed: %s", err.Error())
	}

	h = LogRequestMiddleware(h.ServeHTTP, logger)

	sh := settable.New(h)

	// baseContext for the http server
	baseContext := func(_ net.Listener) context.Context {
		return ctx
	}

	// create the server
	srv := &resetServer{
		Server: http.Server{
			Handler:           LogRequestMiddleware(h.ServeHTTP, logger),
			Addr:              opts.listenAddress,
			ReadHeaderTimeout: 10 * time.Second,
			BaseContext:       baseContext,
		},
		handler: sh,
		logger: logger,
		opts: opts,
	}
	return srv, nil
}

func (s *resetServer) rebuild() error {
	h, err := buildHandler(s.logger, s.opts)
	if err != nil {
		return fmt.Errorf("build hander failed: %s", err.Error())
	}

	h = LogRequestMiddleware(h.ServeHTTP, s.logger)

	// Set the handler
	s.handler.Set(h)

	return nil
}
