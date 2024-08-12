package cache

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/maypok86/otter"
)

type entry struct {
	body       []byte
	header     http.Header
	statusCode int
}

func copyHeader(dst, src http.Header, filter func(k string, v string) bool) {
	if filter == nil {
		filter = func(k string, v string) bool { return true }
	}
	for k, h := range src {
		for _, v := range h {
			if filter(k, v) {
				dst.Add(k, v)
			}
		}
	}
}

func writeEntry(w http.ResponseWriter, e *entry) {
	copyHeader(w.Header(), e.header, nil)
	w.WriteHeader(e.statusCode)
	w.Write(e.body)
}

func CachingHandler(next http.HandlerFunc, size int) http.HandlerFunc {
	logger := slog.Default().With(
		slog.String("handler", "CachingHandler"),
	)
	c, err := otter.MustBuilder[string, *entry](10_000).
		CollectStats().
		Cost(func(key string, value *entry) uint32 {
			return 1
		}).
		WithTTL(time.Hour).
		Build()
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		logger := logger.With(
			slog.String("request.url", r.URL.String()),
		)
		// Check if the key is present.
		if e, found := c.Get(r.URL.String()); found {
			logger.Info("cache hit")
			writeEntry(w, e)
			return
		}

		// Record the response.
		rr := httptest.NewRecorder()
		next(rr, r)
		// Create the entry
		ne := &entry{
			header:     http.Header{},
			statusCode: rr.Result().StatusCode,
		}
		copyHeader(ne.header, rr.Result().Header, nil)
		var err error
		ne.body, err = io.ReadAll(rr.Result().Body)
		if err != nil {
			http.Error(w, "error reading body", http.StatusInternalServerError)
			return
		}

		// write response
		writeEntry(w, ne)

		logger.Info("cache miss",
			slog.String("status", http.StatusText(ne.statusCode)),
			slog.Int("body.len", len(ne.body)),
		)

		// Only cache 2xx results.
		if ne.statusCode < 200 || ne.statusCode >= 300 {
			return
		}

		// Cache the result.
		c.Set(r.URL.String(), ne)
	}
}
