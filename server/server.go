package server

import (
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"sync"

	"github.com/myhops/bbfs"
)

const (
	pathVersions = "/versions"
	pathAll      = "/all"
)

type Server struct {
	serveMux        http.ServeMux
	fsCfg           *bbfs.Config
	rootHandler     http.Handler
	versionHandlers map[string]http.Handler
	logger          *slog.Logger

	tagsMutex sync.Mutex
	tags      []string
}

func (s *Server) GetTags() []string {
	return s.tags
}

func setCacheControlNoCache(header http.Header) {
	header.Set("Cache-Control", "no-cache")
}

func New(
	cfg *bbfs.Config,
	logger *slog.Logger,
	tags []string,
	webFS fs.FS,
	indexTemplate string,
	getInfo func() (*IndexPageInfo, error),
) *Server {

	logger.Info("found tags", slog.Any("tags", tags))

	s := &Server{
		fsCfg:           cfg,
		serveMux:        *http.NewServeMux(),
		versionHandlers: map[string]http.Handler{},
		logger:          logger,
		rootHandler:     http.FileServerFS(bbfs.NewFS(cfg, bbfs.WithLogger(logger))),
		tags:            tags,
	}
	s.routes(webFS, indexTemplate, getInfo)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	setCacheControlNoCache(w.Header())
	s.serveMux.ServeHTTP(w, r)
}

func (s *Server) addVersionRoute(prefix string, tag string) {
	// Create fs with version.
	nfsCfg := *s.fsCfg
	nfsCfg.At = tag
	nfs := bbfs.NewFS(&nfsCfg, bbfs.WithLogger(s.logger))
	// create the pats.
	p, _ := url.JoinPath(prefix, tag, "/")
	// add the handler to the serve mux
	s.serveMux.Handle(fmt.Sprintf("GET %s", p), http.StripPrefix(p, http.FileServerFS(nfs)))
	s.logger.Info("added version handler", "path", p)
}

func (s *Server) addVersionRoutes(prefix string) {
	s.tagsMutex.Lock()
	defer s.tagsMutex.Unlock()
	for _, tag := range s.tags {
		s.addVersionRoute(prefix, tag)
	}
}

func (s *Server) addAllHandler(prefix string) {
	logger := s.logger.With(slog.String("handler", "addAllHandler"))
	nfs := bbfs.NewFS(s.fsCfg, bbfs.WithLogger(s.logger))
	p, _ := url.JoinPath(prefix, "/")
	s.serveMux.Handle(fmt.Sprintf("GET %s", p), http.StripPrefix(p, http.FileServerFS(nfs)))
	logger.Info("added unversioned handler", "path", p)
}

func (s *Server) routes(webFS fs.FS, indexTemplate string, getinfo func() (*IndexPageInfo, error)) {
	// Create the paths for the tags, if any.
	s.addVersionRoutes(pathVersions)
	s.addAllHandler(pathAll)
	s.serveMux.Handle("GET /", s.indexPageHandler(indexTemplate, getinfo))
	s.serveMux.Handle("GET /static/", http.FileServerFS(webFS))
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
func (s *Server) indexPageHandler(tpl string, getInfo func() (*IndexPageInfo, error)) http.Handler {
	logger := s.logger.With(slog.String("handler", "handleIndexPage"))
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
