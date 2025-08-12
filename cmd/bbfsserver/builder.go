package main

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/myhops/bbfs"
	"github.com/myhops/bbfsserver/handlers/cache"
	"github.com/myhops/bbfsserver/resources"
	"github.com/myhops/bbfsserver/server"
)

type builder struct {
	logger *slog.Logger

	opts *options

	bbfsCfg *bbfs.Config
}

// newBuilder constructs a new builder that is not initialized yet.
// To use this builder, call build
func newBuilder(logger *slog.Logger, opts *options) *builder {
	return &builder{
		logger:  logger,
		opts:    opts,
		bbfsCfg: bbfsCfgFromOpts(opts),
	}
}

func (b *builder) build(ctx context.Context) (http.Handler, error) {
	h, err := b.buildHandlerWithMiddleware(ctx)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (b *builder) buildHandlerWithMiddleware(ctx context.Context) (http.Handler, error) {
	bh, err := b.buildHandler(ctx)
	if err != nil {
		return nil, err
	}
	return LogRequestMiddleware(bh.ServeHTTP, b.logger), nil
}

func (b *builder) buildHandler(_ context.Context) (http.Handler, error) {
	allFS := bbfs.NewFS(b.bbfsCfg)

	versions, err := getVersions(b.bbfsCfg, b.logger)
	if err != nil {
		return nil, fmt.Errorf("error getting tags: %w", err)
	}

	tags, err := getTags(b.bbfsCfg, b.logger)
	if err != nil {
		return nil, err
	}

	getinfo := getIndexPageInfo(
		b.opts.repoURL,
		b.opts.title,
		b.bbfsCfg.ProjectKey,
		b.bbfsCfg.RepositorySlug,
		tags,
	)

	webFS, err := fs.Sub(resources.StaticHtmlFS, "web")
	if err != nil {
		return nil, fmt.Errorf("error creating web sub fs: %w", err)
	}

	vfsh := server.New(
		b.logger,
		allFS,
		versions,
		webFS,
		resources.IndexHtmlTemplate,
		getinfo,
		b.opts.changePollingInterval,
		cache.Middleware(10_000),
	)
	return vfsh, nil
}
