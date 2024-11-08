package main

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/myhops/bbfsserver/handlers/sideway"
	"github.com/myhops/bbfsserver/server"

	"github.com/myhops/bbfs"

	"go.uber.org/automaxprocs/maxprocs"
)

func setMaxProcs() {
	pf := func(format string, args ...any) {
		msg := fmt.Sprintf(format, args...)
		slog.Default().Info(msg)
	}
	maxprocs.Set(maxprocs.Logger(pf))
}

func LogRequestMiddleware(next http.HandlerFunc, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.String()
		defer func(start time.Time) {
			passed := time.Since(start)
			logger.Info("handler called", slog.String("url", path), slog.Duration("duration", passed))
		}(time.Now())

		next(w, r)
	}
}

// getIndexPageInfo returns the index pages as html
func getIndexPageInfo(
	bitbucketURL string,
	title string,
	projectKey string,
	repositorySlug string,
	tags []string,
) func() (*server.IndexPageInfo, error) {
	url := &url.URL{
		Path: "/versions",
	}
	var versions []struct {
		Name string
		Path string
	}
	for _, tag := range tags {
		parts := strings.Split(tag, "/")
		module := ""
		if len(parts) == 2 {
			module = parts[0]
		}
		v := struct {
			Name string
			Path string
		}{
			Name: tag,
			Path: url.JoinPath(tag, module, "/").String(),
		}
		versions = append(versions, v)
	}

	return func() (*server.IndexPageInfo, error) {
		res := &server.IndexPageInfo{
			BitbucketURL:   bitbucketURL,
			Title:          title,
			ProjectKey:     projectKey,
			RepositorySlug: repositorySlug,
			Versions:       versions,
		}
		return res, nil
	}
}

func latestTagChanged(lastTag string, cfg *bbfs.Config, logger *slog.Logger) bool {
	logger = logger.With(slog.String("method", "main.latestTagChanged"))
	t := getLatestTag(cfg, logger)
	if t == "" {
		logger.Info("no tag found")
		return false
	}
	changed := t != "" && t != lastTag
	if changed {
		logger.Info("new tag found",
			slog.String("lastTag", lastTag),
			slog.String("newTag", t))
	}
	return changed
}

func getDryRunVersions(cfg *bbfs.Config, logger *slog.Logger) []*server.Version {
	tags := []string{"testtag1", "testtag2/v1"}
	res, _ := getVersionsFromTags(cfg, logger, tags)
	return res
}

func bbfsCfgFromOpts(opts *options) *bbfs.Config {
	return &bbfs.Config{
		Host:           opts.host,
		ProjectKey:     opts.projectKey,
		RepositorySlug: opts.repositorySlug,
		AccessKey:      opts.accessKey,
	}
}

func runWithOpts(ctx context.Context, logger *slog.Logger, opts *options) error {
	// create context that catches kill and interrupt
	ctx, stop := signal.NotifyContext(ctx, os.Kill, os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Create a handler that sends a signal to a channel to trigger a rebuild
	rebuildChan := make(chan struct{}, 1)
	rebuildhandler := func(w http.ResponseWriter, r *http.Request) {
		logger := logger.With(slog.String("method", "rebuildHandler"))
		logger.Info("rebuild requested")
		// Send a signal but fail if queue is full
		select {
		case rebuildChan <- struct{}{}:
			logger.Info("sent signal to trigger rebuild")
		default:
			logger.Info("could not send signal to trigger requild")
		}
	}

	// Build the rebuild handler
	rebuildHandler, err := newRebuildHandler(ctx, logger, opts)
	if err != nil {
		return err
	}

	// Add a callback for rebuild
	sidewayHandler := sideway.New(rebuildHandler, logger)
	sidewayHandler.HandleFunc("/api/controllers/rebuild", rebuildhandler)
	
	// build the server
	srv, err := newRebuildServer(ctx, logger, opts, sidewayHandler, rebuildHandler.Rebuild)
	if err != nil {
		return fmt.Errorf("error building server: %s", err.Error())
	}

	// Start the server in the background
	go func() {
		logger := logger.With("goroutine", "listen and serve")
		logger.Info("starting server")
		if err := srv.ListenAndServe(); err != nil {
			logger.Error("error", "error", err.Error())
		}
		logger.Info("server stopped")
	}()

	rebuild := func(msg string) {
		logger := logger.With(slog.String("message", msg))
		cfg := bbfsCfgFromOpts(opts)
		if !latestTagChanged(srv.latestTag, cfg, logger) {
			logger.Info("no changes detected")
			return
		}
		logger.Info("changes detected")
		// rebuild the server
		logger.Info("start server rebuild")
		if err := srv.rebuild(ctx); err != nil {
			logger.Error("error rebuilding server", slog.String("error", err.Error()))
		}
	}

FOR:
	for {
		select {
		case <-ctx.Done():
			break FOR
		case <-time.After(opts.changePollingInterval):
			rebuild("timer triggered")
		case <-rebuildChan:
			rebuild("rebuild callback")
		}
	}

	// Wait for a signal
	<-ctx.Done()
	log.Print("Done closed")
	if ctx.Err() != nil {
		log.Printf("error: %s", ctx.Err().Error())
	}

	// shutdown the server and wait for 10 seconds
	sctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = srv.Shutdown(sctx)
	if err != nil {
		return fmt.Errorf("shutdown failed: %w", err)
	}
	log.Print("server shut down normally")
	return nil
}

// Run runs the program.
func run(
	ctx context.Context,
	args []string,
	getenv func(string) string,
	stderr io.Writer,
) error {
	if len(args) > 1 && args[1] == "-h" {
		fmt.Println(usageText)
		return nil
	}
	opts := defaultOptions()
	opts.fromEnv(getenv)

	initLogger(opts.logFormat, stderr)

	logger := slog.Default()

	// set the max procs
	setMaxProcs()

	logger.Info("options are",
		"host", opts.host,
		"listenAddress", opts.listenAddress,
		"projectKey", opts.projectKey,
		"repositorySlug", opts.repositorySlug,
		slog.Duration("pollingInterval", opts.changePollingInterval),
	)
	return runWithOpts(ctx, logger, opts)
}

func initLogger(logFormat string, lw io.Writer) {
	var lh slog.Handler
	ho := &slog.HandlerOptions{}
	switch strings.ToLower(logFormat) {
	case "text":
		lh = slog.NewTextHandler(lw, ho)
	default:
		lh = slog.NewJSONHandler(lw, ho)
	}
	logger := slog.New(lh)
	slog.SetDefault(logger)
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered error in main: %v\nStack trace:\n%s", r, string(debug.Stack()))
		}
	}()
	err := run(context.Background(), os.Args, os.Getenv, os.Stderr)
	if err != nil {
		log.Printf("run error: %s", err.Error())
	}
}
