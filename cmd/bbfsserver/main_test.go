package main

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/myhops/bbfs"
	"github.com/myhops/bbfsserver/handlers/cache"
	"github.com/myhops/bbfsserver/resources"
	"github.com/myhops/bbfsserver/server"
)

func testGetOptionsFromEnvGetenv(key string) string {
	switch key {
	case "PORT":
		return "10000"
	case "BBFSSRV_HOST":
		return "BBHOST.example.com"
	case "BBFSSRV_PROJECT_KEY":
		return "projectKey"
	case "BBFSSRV_REPOSITORY_SLUG":
		return "repoSlug"
	case "BBFSSRV_ACCESS_KEY":
		return "accessKey"
	case "BBFSSRV_LOG_FORMAT":
		return "json"
	case "BBFSSRV_TAG_POLL_INTERVAL":
		return "1s"
	default:
		return ""
	}
}

func dryRunGetOptionsFromEnvGetenv(key string) string {
	switch key {
	case "PORT":
		return "10000"
	case "BBFSSRV_HOST":
		return "BBHOST.example.com"
	case "BBFSSRV_PROJECT_KEY":
		return "projectKey"
	case "BBFSSRV_REPOSITORY_SLUG":
		return "repoSlug"
	case "BBFSSRV_ACCESS_KEY":
		return "accessKey"
	case "BBFSSRV_LOG_FORMAT":
		return "json"
	case "BBFSSRV_TAG_POLL_INTERVAL":
		return "1s"
	case "BBFSSRV_DRY_RUN":
		return "true"
	default:
		return ""
	}
}

func TestGetOptionsFromEnv(t *testing.T) {
	opts := &options{}
	opts.fromEnv(testGetOptionsFromEnvGetenv)

	if opts.changePollingInterval != time.Second {
		t.Errorf("want %v, got %v", time.Minute, opts.changePollingInterval)
	}
}

func TestDryRun(t *testing.T) {
	opts := &options{}
	opts.fromEnv(dryRunGetOptionsFromEnvGetenv)
	cfg := &bbfs.Config{}
	out := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{}))
	allFS := bbfs.NewFS(cfg)
	versions := getDryRunVersions(cfg, logger)
	getinfo := getIndexPageInfo("repoURL", "Title", "Project 1", "Repo 1", []string{"tag1"})
	h := server.New(
		logger, 
		allFS, 
		versions, 
		resources.StaticHtmlFS, 
		resources.IndexHtmlTemplate, 
		getinfo, opts.changePollingInterval,
		cache.Middleware(10_000),
	)

	srv := httptest.NewServer(h)
	defer srv.Close()
	u := srv.URL
	_ = u

	r, err := http.Get(srv.URL)
	if err != nil {
		t.Errorf("error getting page: %s", err.Error())
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Errorf("error reading body: %s", err.Error())
	}
	bodys := string(body)
	_ = bodys
	t.Logf("status: %s", r.Status)
}

func TestIndexPage(t *testing.T) {
	opts := &options{}
	opts.fromEnv(dryRunGetOptionsFromEnvGetenv)
	cfg := &bbfs.Config{}
	out := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{}))
	allFS := bbfs.NewFS(cfg)
	versions := getDryRunVersions(cfg, logger)
	getinfo := getIndexPageInfo("repoURL", "Title", "Project 1", "Repo 1", []string{"tag1"})
	srv := server.New(logger, 
		allFS, 
		versions, 
		resources.StaticHtmlFS, 
		resources.IndexHtmlTemplate, 
		getinfo, opts.changePollingInterval,
		cache.Middleware(10_000))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, r)
	body, err := io.ReadAll(w.Result().Body)
	if err != nil {
		t.Errorf("error reading body: %s", err.Error())
	}
	bodys := string(body)
	_ = bodys
	t.Logf("status: %s", w.Result().Status)
}
