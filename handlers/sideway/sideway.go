package sideway

import (
	"log/slog"
	"net/http"

	"github.com/myhops/bbfs/nulllog"
)

type Handler struct {
	handler *http.ServeMux

	next   http.Handler
	logger *slog.Logger
}

func New(next http.Handler, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = nulllog.Logger()
	}
	return &Handler{
		handler: http.NewServeMux(),
		next:    next,
		logger: logger,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := h.logger.With(slog.String("method", "sideway.ServeHTTP"))
	mh, pattern := h.handler.Handler(r)
	if pattern != "" {
		logger.Info("calling sideway", slog.String("pattern", pattern))
		mh.ServeHTTP(w,r)
		return
	}
	logger.Info("calling next")
	h.next.ServeHTTP(w,r)
}

func (h *Handler) Handle(pattern string, handler http.Handler) {
	h.handler.Handle(pattern, handler)
}

func (h *Handler) HandleFunc(pattern string, handler http.HandlerFunc) {
	h.handler.HandleFunc(pattern, handler)
}
