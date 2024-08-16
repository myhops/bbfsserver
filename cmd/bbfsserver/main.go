package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
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

	"github.com/myhops/bbfsserver/cache"
	"github.com/myhops/bbfsserver/handlers"

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

//go:embed resources/web/index.html
var indexHtmlTemplate string

//go:embed resources/web
var staticHtmlFS embed.FS

type options struct {
	host             string
	logFormat        string
	listenAddress    string
	projectKey       string
	repositorySlug   string
	accessKey        string
	tagsPollInterval time.Duration
	dryRun           string
}

func defaultOptions() *options {
	return &options{
		logFormat:     "json",
		listenAddress: ":8080",
	}
}

func setIfSet(v string, val *string) {
	if v != "" {
		*val = v
	}
}

func compareTags(t1, t2 []string) int {
	slices.Sort(t1)
	slices.Sort(t2)
	return slices.Compare(t1, t2)
}

func getPollInterval(interval string) time.Duration {
	res := newTagPollingInterval
	if interval == "" {
		return res
	}
	i, err := time.ParseDuration(interval)
	if err != nil {
		return res
	}
	if i < time.Second {
		return time.Second
	}
	res = i
	return res
}

func (o *options) fromEnv(getenv func(string) string) {
	setIfSet(getenv("PORT"), &o.listenAddress)
	setIfSet(getenv("BBFSSRV_LISTEN_ADDRESS"), &o.listenAddress)
	setIfSet(getenv("BBFSSRV_HOST"), &o.host)
	setIfSet(getenv("BBFSSRV_PROJECT_KEY"), &o.projectKey)
	setIfSet(getenv("BBFSSRV_REPOSITORY_SLUG"), &o.repositorySlug)
	setIfSet(getenv("BBFSSRV_ACCESS_KEY"), &o.accessKey)
	setIfSet(getenv("BBFSSRV_LOG_FORMAT"), &o.logFormat)
	setIfSet(getenv("BBFSSRV_DRY_RUN"), &o.dryRun)

	o.tagsPollInterval = getPollInterval(getenv("BBFSSRV_TAG_POLL_INTERVAL"))

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

// getIndexPageInfo returns the
func getIndexPageInfo(
	title string,
	projectKey string,
	repositorySlug string,
) func() (*IndexPageInfo, error) {
	return func() (*IndexPageInfo, error) {
		res := &IndexPageInfo{
			Title:          title,
			ProjectKey:     projectKey,
			RepositorySlug: repositorySlug,
			Versions: []struct {
				Name string
				Path string
			}{
				{
					Name: "tag1",
					Path: "https://www.google.com",
				},
				{
					Name: "tag2",
					Path: "https://www.booking.com",
				},
				{
					Name: "tag1",
					Path: "https://www.google.com",
				},
				{
					Name: "tag2",
					Path: "https://www.booking.com",
				},
				{
					Name: "tag1",
					Path: "https://www.google.com",
				},
				{
					Name: "tag2",
					Path: "https://www.booking.com",
				},
				{
					Name: "tag1",
					Path: "https://www.google.com",
				},
				{
					Name: "tag2",
					Path: "https://www.booking.com",
				},
			},
		}
		return res, nil
	}
}

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
		slog.Duration("pollingInterval", opts.tagsPollInterval),
	)

	cfg := &bbfs.Config{
		Host:           opts.host,
		ProjectKey:     opts.projectKey,
		RepositorySlug: opts.repositorySlug,
		AccessKey:      opts.accessKey,
	}

	getinfo := getIndexPageInfo(
		"OLO KOR Build Reports",
		opts.projectKey,
		opts.repositorySlug,
	)

	tags := []string{"testtag1", "testtag2"}
	if opts.dryRun != "true" {
		t, err := getTags(cfg, logger)
		if err != nil {
			return fmt.Errorf("error getting tags: %w", err)
		}
		tags = t
	}

	webFS, err := fs.Sub(staticHtmlFS, "resources/web")
	if err != nil {
		return fmt.Errorf("error creating resources/web sub fs: %w", err)
	}

	vfsh := newVersionFileServerFS(cfg, logger, tags, webFS, indexHtmlTemplate, getinfo)
	settableVfsh := handlers.NewSettable(cache.CachingHandler(vfsh.ServeHTTP, 10_000))

	// create context that catches kill and interrupt
	ctx, stop := signal.NotifyContext(context.Background(), os.Kill, os.Interrupt, syscall.SIGTERM)
	defer stop()

	// baseContext for the http server
	baseContext := func(_ net.Listener) context.Context {
		return ctx
	}

	// create the server
	server := http.Server{
		Handler:           LogRequestMiddleware(settableVfsh.ServeHTTP, logger),
		Addr:              opts.listenAddress,
		ReadHeaderTimeout: 10 * time.Second,
		BaseContext:       baseContext,
	}
	// Start the server in the background
	go func() {
		logger := logger.With("goroutine", "listen and serve")
		logger.Info("starting server")
		if err := server.ListenAndServe(); err != nil {
			logger.Error("error", "error", err.Error())
		}
		logger.Info("server stopped")
	}()

	// Add tag checker.
	go func() {
		logger := logger.With("goroutine", "tag checker")
		// Check every 5 minutes
		tick := time.NewTicker(opts.tagsPollInterval)
		for {
			select {
			case <-ctx.Done():
				logger.Info("done received")
				return
			case <-tick.C:
				logger.Info("checking for new tags")
				t1, err := getTags(cfg, logger)
				if err != nil {
					break
				}
				if compareTags(t1, vfsh.getTags()) == 0 {
					break
				}
				vfsh = newVersionFileServerFS(cfg, logger, t1, webFS, indexHtmlTemplate, getinfo)
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
	err = server.Shutdown(sctx)
	if err != nil {
		return fmt.Errorf("shutdown failed: %w", err)
	}
	log.Print("server shut down normally")
	return nil
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
			log.Printf("Recovered error in main: %v", r)
		}
	}()
	err := run(context.Background(), os.Args, os.Getenv, os.Stderr)
	if err != nil {
		log.Printf("run error: %s", err.Error())
	}
}
