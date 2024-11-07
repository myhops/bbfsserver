package rebuild

import (
	"context"
	"fmt"
	"net/http"

	"github.com/myhops/bbfsserver/handlers/settable"
)

var ErrHandlerNotSet = fmt.Errorf("handler not set")

type RebuildHandler struct {
	// BuildHandler builds a new handler.
	BuildHandler func(context.Context) (http.Handler, error)

	handler settable.Settable
}

// NewNoRebuild creates a new RebuildHandler, but does not build the handler.
// This allows you use a different context for the rebuild than for the New
func NewNoRebuild(bh func(context.Context) (http.Handler, error)) *RebuildHandler {
	h := &RebuildHandler{
		BuildHandler: bh,
	}
	return h
}

// New creates and builds a new handler.
// The context is used during the rebuild.
func New(ctx context.Context, bh func(context.Context) (http.Handler, error)) (*RebuildHandler, error) {
	h := NewNoRebuild(bh)
	if err := h.rebuild(ctx); err != nil {
		return nil, err
	}
	return h, nil	
}

func (h *RebuildHandler) rebuild(ctx context.Context) error {
	if h.BuildHandler == nil {
		return ErrHandlerNotSet
	}
	nh, err := h.BuildHandler(ctx)
	if err != nil {
		return err
	}
	h.handler.Set(nh)
	return nil
}

// Rebuild rebuilds and sets the handler.
func (h *RebuildHandler) Rebuild(ctx context.Context) error {
	return h.rebuild(ctx)
}

// ServeHTTP passes the request to the handler that BuildHandler creates.
func (h *RebuildHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}