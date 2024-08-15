package main

import (
	"context"
	"fmt"
	"html/template"
	"io/fs"
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

func newVersionFileServerFS(
	cfg *bbfs.Config,
	logger *slog.Logger,
	tags []string,
	webFS fs.FS,
	indexTemplate string,
	getInfo func() (*IndexPageInfo, error),
) *versionFileServerFS {

	logger.Info("found tags", slog.Any("tags", tags))

	h := &versionFileServerFS{
		fsCfg:           cfg,
		serveMux:        *http.NewServeMux(),
		versionHandlers: map[string]http.Handler{},
		logger:          logger,
		rootHandler:     http.FileServerFS(bbfs.NewFS(cfg, bbfs.WithLogger(logger))),
		tags:            tags,
	}
	h.routes(webFS, indexTemplate, getInfo)

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
	logger := h.logger.With(slog.String("handler", "addAllHandler"))
	nfs := bbfs.NewFS(h.fsCfg, bbfs.WithLogger(h.logger))
	p, _ := url.JoinPath(prefix, "/")
	h.serveMux.Handle(fmt.Sprintf("GET %s", p), http.StripPrefix(p, http.FileServerFS(nfs)))
	logger.Info("added unversioned handler", "path", p)
}

func (h *versionFileServerFS) routes(webFS fs.FS, indexTemplate string, getinfo func() (*IndexPageInfo, error)) {
	// Create the paths for the tags, if any.
	h.addVersionRoutes(pathVersions)
	h.addAllHandler(pathAll)
	h.serveMux.Handle("GET /", h.indexPageHandler(indexTemplate, getinfo))
	h.serveMux.Handle("GET /web/", http.FileServerFS(webFS))
}

type IndexPageInfo struct {
	Title          string
	BitbucketURL   string
	ProjectKey     string
	RepositorySlug string
	Versions       []struct {
		Name string
		Path string
	}
}

// handleIndexPage shows a welcome page with
func (h *versionFileServerFS) indexPageHandler(tpl string, getInfo func() (*IndexPageInfo, error)) http.Handler {
	logger := h.logger.With(slog.String("handler", "handleIndexPage"))
	// Load the template file and parse it
	t := template.New("index")
	t, err := t.Parse(tpl)
	if err != nil {
		logger.Error("template parsing failed",
			slog.String("error", err.Error()),
		)
	}

	f := func(w http.ResponseWriter, r *http.Request) {
		if t == nil {
			logger.Error("template not available")
			http.Redirect(w, r, "/all/", http.StatusPermanentRedirect)
			return
		}
		info, err := getInfo()
		if err != nil {
			logger.Error("error getting info", slog.String("error", err.Error()))
			http.Error(w, fmt.Sprintf("error getting info: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		err = t.Execute(w, info)
		if err != nil {
			logger.Error("error executing template",
				slog.String("error", err.Error()),
				slog.Any("info", info),
			)
		}
	}

	return http.HandlerFunc(f)
}
