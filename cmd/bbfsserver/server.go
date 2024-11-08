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
	handler *rebuild.RebuildHandler

	latestTag string
	bbfsCfg   *bbfs.Config
	logger    *slog.Logger
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

func newServer(ctx context.Context, logger *slog.Logger, opts *options) (*rebuildServer, error) {
	// baseContext for the http server
	baseContext := func(_ net.Listener) context.Context {
		return ctx
	}

	// Create the builder.
	builder := newBuilder(logger, opts)
	// Create the rebuild handler.
	handler, err := rebuild.New(ctx, builder.build)
	if err != nil {
		return nil, err
	}

	srv := &rebuildServer{
		Server: http.Server{
			Addr:              opts.listenAddress,
			ReadHeaderTimeout: 10 * time.Second,
			BaseContext:       baseContext,
			Handler:           handler,
		},
		handler:   handler,
		latestTag: getLatestTag(bbfsCfgFromOpts(opts), logger),
		bbfsCfg:   bbfsCfgFromOpts(opts),
		logger:    logger,
	}

	return srv, nil
}

func (s *rebuildServer) rebuild(ctx context.Context) error {
	// Save the latest tag
	s.latestTag = getLatestTag(s.bbfsCfg, s.logger)
	return s.handler.Rebuild(ctx)
}
