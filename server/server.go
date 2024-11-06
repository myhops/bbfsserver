package server

import (
	"fmt"
	"html/template"
	"io/fs"
	"iter"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	pathVersions = "/versions"
	pathAll      = "/all"
)

type Version struct {
	Name string
	Dir  fs.FS
}

type Server struct {
	serveMux http.ServeMux
	logger   *slog.Logger
	all      fs.FS
	versions []*Version

	// timeToLive
	ttlMutex   sync.RWMutex
	timeToLive time.Duration
	startTime  time.Time

	cacheMiddleware func(next http.Handler) http.Handler
}

// Tags returns an iterator, go 1.23.0, just for the fun of it.
// func (s *Server) Versions() func(yield func(string) bool) bool {
func (s *Server) Versions() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, v := range s.versions {
			if !yield(v.Name) {
				return
			}
		}
	}
}

// GetVersionNames returns an array with the prefixes of the tags
func (s *Server) GetVersionNames() []string {
	res := make([]string, 0, len(s.versions))
	for _, t := range s.versions {
		res = append(res, t.Name)
	}
	return res
}

// New creates a new server using
func New(
	// logger
	logger *slog.Logger,
	// all is the FS for the main branch
	all fs.FS,
	// versions is a list of Version, which contain the name of the ref and the FS
	versions []*Version,
	// webFS is the FS for the static files
	webFS fs.FS,
	// indexTemplate is the http/template for index.html
	indexTemplate string,
	// getInfo is a function that returns the struct that indexTemplate uses
	getInfo func() (*IndexPageInfo, error),
	// timeToLive sets the time the server is expected to run
	timeToLive time.Duration,
	// cacheMiddleware caches requests based on the path of the request
	cacheMiddleware func(next http.Handler) http.Handler,
) *Server {
	s := &Server{
		serveMux:        *http.NewServeMux(),
		logger:          logger,
		all:             all,
		versions:        versions,
		timeToLive:      timeToLive,
		startTime:       time.Now(),
		cacheMiddleware: cacheMiddleware,
	}
	s.routes(webFS, indexTemplate, getInfo)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.setCacheControl(w.Header())
	s.serveMux.ServeHTTP(w, r)
}

func (s *Server) addPrefixFSRoute(prefix string, version *Version) {
	// create the path.
	p, _ := url.JoinPath(prefix, "/", version.Name, "/")
	// add the handler to the serve mux
	s.serveMux.Handle(fmt.Sprintf("GET %s", p), s.cacheMiddleware(http.StripPrefix(p, http.FileServerFS(version.Dir))))
	s.logger.Info("added prefix FS", "path", p)
}

func (s *Server) addVersionRoutes(prefix string) {
	for _, p := range s.versions {
		s.addPrefixFSRoute(prefix, p)
	}
}

func (s *Server) addAllRoute(prefix string, fs fs.FS) {
	logger := s.logger.With(slog.String("handler", "addAllHandler"))
	p, _ := url.JoinPath(prefix, "/")
	s.serveMux.Handle(fmt.Sprintf("GET %s", p), s.cacheMiddleware(http.StripPrefix(p, http.FileServerFS(fs))))
	logger.Info("added unversioned handler", "path", p)
}

func (s *Server) routes(
	webFS fs.FS,
	indexTemplate string,
	getinfo func() (*IndexPageInfo, error),
) {
	// Create the paths for the tags, if any.
	s.addVersionRoutes(pathVersions)
	s.addAllRoute(pathAll, s.all)
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

func (s *Server) setCacheControl(header http.Header) {
	const cacheControl = "Cache-Control"

	logger := s.logger.With(slog.String("server.method", "setCacheControl"))
	s.ttlMutex.RLock()
	maxAge := s.timeToLive - time.Since(s.startTime)
	s.ttlMutex.RUnlock()
	if maxAge < 0 {
		maxAge = 0
	}
	val := fmt.Sprintf("max-age=%d", int64(maxAge.Seconds()))
	header.Set(cacheControl, val)
	logger.Info("set cache control", slog.String(cacheControl, val))
}

// func (s *Server) ResetStartTime() {
// 	logger := s.logger.With(slog.String("server.method", "ResetStartTime"))

// 	s.ttlMutex.Lock()
// 	s.startTime = time.Now()
// 	s.ttlMutex.Unlock()

// 	logger.Info("start time reset", slog.Time("startTime", s.startTime))
// }
