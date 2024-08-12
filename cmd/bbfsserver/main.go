package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	"bbfsserver/cache"
	"bbfsserver/handlers"

	"github.com/myhops/bbfs"

	"go.uber.org/automaxprocs/maxprocs"
)

const (
	newTagPollingInterval = 5 * time.Minute
)

func setMaxProcs() {
	pf := func(format string, args ...any) {
		msg := fmt.Sprintf(format, args...)
		slog.Default().Info(msg)
	}
	maxprocs.Set(maxprocs.Logger(pf))
}

//go:embed resources/usage.txt
var usageText string

type options struct {
	host             string
	logFormat        string
	listenAddress    string
	projectKey       string
	repositorySlug   string
	accessKey        string
	tagsPollInterval time.Duration
}

func defaultOptions() *options {
	return &options{
		logFormat:     "json",
		listenAddress: ":8080",
	}
}

func setFromEnv(key string, val *string) {
	if v := os.Getenv(key); v != "" {
		*val = v
	}
}

func compareTags(t1, t2 []string) int {
	slices.Sort(t1)
	slices.Sort(t2)
	return slices.Compare(t1, t2)
}

func getPollInterval(interval string) time.Duration {
	res := 5*time.Minute
	if interval == "" {
		return res
	}
	i, err := time.ParseDuration(interval);
	if err != nil {
		return res
	}
	if i < time.Second {
		return time.Second
	}
	res = i
	return res
}

func (o *options) fromEnv() {
	setFromEnv("PORT", &o.listenAddress)
	setFromEnv("BBFSSRV_LISTEN_ADDRESS", &o.listenAddress)
	setFromEnv("BBFSSRV_HOST", &o.host)
	setFromEnv("BBFSSRV_PROJECT_KEY", &o.projectKey)
	setFromEnv("BBFSSRV_REPOSITORY_SLUG", &o.repositorySlug)
	setFromEnv("BBFSSRV_ACCESS_KEY", &o.accessKey)
	setFromEnv("BBFSSRV_LOG_FORMAT", &o.logFormat)

	o.tagsPollInterval = getPollInterval(os.Getenv("BBFSSRV_TAG_POLL_INTERVAL"))
	
	// fix listen address if needed.
	if o.listenAddress[0] != ':' {
		o.listenAddress = ":" + o.listenAddress
	}
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

func run(args []string) error {
	if len(args) > 1 && args[1] == "-h" {
		fmt.Println(usageText)
		return nil
	}
	opts := defaultOptions()
	opts.fromEnv()

	initLogger(opts.logFormat)

	logger := slog.Default()

	// set the max procs
	setMaxProcs()

	logger.Info("options are",
		"host", opts.host,
		"listenAddress", opts.listenAddress,
		"projectKey", opts.projectKey,
		"repositorySlug", opts.repositorySlug,
		slog.Duration("pollingInterval", opts.tagsPollInterval),
	)

	cfg := &bbfs.Config{
		Host:           opts.host,
		ProjectKey:     opts.projectKey,
		RepositorySlug: opts.repositorySlug,
		AccessKey:      opts.accessKey,
	}

	vfsh := newVersionFileServerFS(cfg, logger)
	settableVfsh := handlers.NewSettable(vfsh)

	// create context that catches kill and interrupt
	ctx, stop := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt, syscall.SIGTERM)
	defer stop()

	// baseContext for the http server
	baseContext := func(_ net.Listener) context.Context {
		return ctx
	}

	// create the server
	server := http.Server{
		Handler:           LogRequestMiddleware(cache.CachingHandler(settableVfsh.ServeHTTP, 10_000), logger),
		Addr:              opts.listenAddress,
		ReadHeaderTimeout: 10 * time.Second,
		BaseContext:       baseContext,
	}
	// Start the server in the background
	go func() {
		logger := logger.With("goroutine", "listen and serve")
		if err := server.ListenAndServe(); err != nil {
			logger.Error("error", "error", err.Error())
		}
		logger.Info("server stopped")
	}()

	// Add tag checker.
	go func() {
		logger := logger.With("goroutine", "tag checker")
		// Check every 5 minutes
		tick := time.NewTicker(newTagPollingInterval)
		for {
			select {
			case <-ctx.Done():
				logger.Info("done received")
				return
			case <-tick.C:
				logger.Info("")
				t1, err := getTags(cfg, logger)
				if err != nil {
					break
				}
				if compareTags(t1, vfsh.getTags()) == 0 {
					break
				}
				vfsh = newVersionFileServerFS(cfg, logger)
				settableVfsh.Set(vfsh)
			}
		}
	}()

	// Wait for a signal
	<-ctx.Done()
	log.Print("Done closed")
	if ctx.Err() != nil {
		log.Printf("error: %s", ctx.Err().Error())
	}

	// shutdown the server and wait for 10 seconds
	sctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := server.Shutdown(sctx)
	if err != nil {
		return fmt.Errorf("shutdown failed: %w", err)
	}
	log.Print("server shut down normally")
	return nil
}

func initLogger(logFormat string) {
	var lh slog.Handler
	ho := &slog.HandlerOptions{}
	lw := os.Stderr
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
	run(os.Args)
}
