package main

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/myhops/bbfs"
)

func TestIndexPage(t *testing.T) {
	out := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{}))

	tags := []string{"tag1", "tag2"}
	cfg := &bbfs.Config{}
	getinfo := getIndexPageInfo("Title", "Project 1", "Repo 1")
	srv := newVersionFileServerFS(cfg, logger, tags, staticHtmlFS, indexHtmlTemplate, getinfo)
	h := srv.indexPageHandler(indexHtmlTemplate, getinfo)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	body, err := io.ReadAll(w.Result().Body)
	if err != nil {
		t.Errorf("error reading body: %s", err.Error())
	}
	bodys := string(body)
	_ = bodys
	t.Logf("status: %s", w.Result().Status)
}

func TestIndexPageWithServer(t *testing.T) {
	out := &bytes.Buffer{}
	logger := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{}))

	tags := []string{"tag1", "tag2"}
	cfg := &bbfs.Config{}
	getinfo := getIndexPageInfo("Title", "Project 1", "Repo 1")
	h := newVersionFileServerFS(cfg, logger, tags, staticHtmlFS, indexHtmlTemplate, getinfo)
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

