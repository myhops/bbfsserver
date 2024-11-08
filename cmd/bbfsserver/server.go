package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/myhops/bbfs"
	"github.com/myhops/bbfsserver/handlers/rebuild"
)

type rebuildServer struct {
	http.Server
	handler http.Handler

	rebuildFunc func(context.Context) error

	latestTag string
	bbfsCfg   *bbfs.Config
	logger    *slog.Logger
}

type rebuildServerOption func(s *rebuildServer)

func WithMiddleware(func(next http.Handler) http.Handler) rebuildServerOption {
	return func(s *rebuildServer) {

	}
}

func getLatestTag(cfg *bbfs.Config, logger *slog.Logger) string {
	tags, err := getTags(cfg, logger)
	if err != nil {
		return ""
	}
	if len(tags) == 0 {
		return ""
	}
	return tags[0]
}

func newRebuildServer(
	ctx context.Context,
	logger *slog.Logger,
	opts *options,
	handler http.Handler,
	rebuildFunc func(context.Context) error,
) (*rebuildServer, error) {
	// baseContext for the http server
	baseContext := func(_ net.Listener) context.Context {
		return ctx
	}

	bbfsCfg := bbfsCfgFromOpts(opts)
	latestTag := getLatestTag(bbfsCfg, logger)
	srv := &rebuildServer{
		Server: http.Server{
			Addr:              opts.listenAddress,
			ReadHeaderTimeout: 10 * time.Second,
			BaseContext:       baseContext,
			Handler:           handler,
		},
		handler:   handler,
		latestTag: latestTag,
		bbfsCfg:   bbfsCfg,
		logger:    logger,
	}

	return srv, nil
}

func newRebuildHandler(ctx context.Context, logger *slog.Logger, opts *options) (*rebuild.RebuildHandler, error) {
	// Create the builder.
	builder := newBuilder(logger, opts)
	// Create the rebuild handler.
	handler, err := rebuild.New(ctx, builder.build)
	if err != nil {
		return nil, err
	}
	return handler, nil
}

func (s *rebuildServer) rebuild(ctx context.Context) error {
	// Save the latest tag
	s.latestTag = getLatestTag(s.bbfsCfg, s.logger)
	return s.rebuildFunc(ctx)
}
