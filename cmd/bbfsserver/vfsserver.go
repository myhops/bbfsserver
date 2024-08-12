package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"sync"

	"github.com/myhops/bbfs"
	"github.com/myhops/bbfs/bbclient/server"
)

const (
	pathVersions = "/versions"
	pathAll      = "/all"
)

type versionFileServerFS struct {
	serveMux        http.ServeMux
	fsCfg           *bbfs.Config
	rootHandler     http.Handler
	versionHandlers map[string]http.Handler
	logger          *slog.Logger

	tagsMutex sync.Mutex
	tags      []string
}

func getTags(cfg *bbfs.Config, logger *slog.Logger) ([]string, error) {
	u := url.URL{
		Scheme: "https",
		Host:   cfg.Host,
		Path:   filepath.Join(bbfs.ApiPath, bbfs.DefaultVersion),
	}

	// Find the valid tags
	client := server.Client{
		BaseURL:   u.String(),
		AccessKey: server.SecretString(cfg.AccessKey),
		Logger:    logger,
	}
	tags, err := client.GetTags(context.Background(), &server.GetTagsCommand{
		ProjectKey: cfg.ProjectKey,
		RepoSlug:   cfg.RepositorySlug,
	})
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (h *versionFileServerFS) getTags() []string {
	return h.tags
}

func setCacheControlNoCache(header http.Header) {
	header.Set("Cache-Control", "no-cache")
}

func newVersionFileServerFS(cfg *bbfs.Config, logger *slog.Logger) *versionFileServerFS {
	tags, err := getTags(cfg, logger)
	if err != nil {
		panic(fmt.Sprintf("error getting tags %s", err.Error()))
	}

	logger.Info("found tags", slog.Any("tags", tags))

	h := &versionFileServerFS{
		fsCfg:           cfg,
		serveMux:        *http.NewServeMux(),
		versionHandlers: map[string]http.Handler{},
		logger:          logger,
		rootHandler:     http.FileServerFS(bbfs.NewFS(cfg, bbfs.WithLogger(logger))),
		tags:            tags,
	}
	h.routes()

	return h
}

func (h *versionFileServerFS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setCacheControlNoCache(w.Header())
	h.serveMux.ServeHTTP(w, r)
}

func (h *versionFileServerFS) addVersionRoute(prefix string, tag string) {
	// Create fs with version.
	nfsCfg := *h.fsCfg
	nfsCfg.At = tag
	nfs := bbfs.NewFS(&nfsCfg, bbfs.WithLogger(h.logger))
	// create the path.
	p, _ := url.JoinPath(prefix, tag, "/")
	// add the handler to the serve mux
	h.serveMux.Handle(fmt.Sprintf("GET %s", p), http.StripPrefix(p, http.FileServerFS(nfs)))
	h.logger.Info("added version handler", "path", p)
}

func (h *versionFileServerFS) addVersionRoutes(prefix string) {
	h.tagsMutex.Lock()
	defer h.tagsMutex.Unlock()
	for _, tag := range h.tags {
		h.addVersionRoute(prefix, tag)
	}
}

func (h *versionFileServerFS) addAllHandler(prefix string) {
	nfs := bbfs.NewFS(h.fsCfg, bbfs.WithLogger(h.logger))
	p, _ := url.JoinPath(prefix, "/")
	h.serveMux.Handle(fmt.Sprintf("GET %s", p), http.StripPrefix(p, http.FileServerFS(nfs)))
	h.logger.Info("added unversioned handler", "path", p)
}

func (h *versionFileServerFS) routes() {
	// Create the paths for the tags, if any.
	h.addVersionRoutes(pathVersions)
	h.addAllHandler(pathAll)
	h.serveMux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		logger := h.logger.With("handler", "allRedirector")
		logger.Info("called")
		http.Redirect(w, r, "/all/", http.StatusMovedPermanently)
	})
}
