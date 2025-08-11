package settable

import (
	"net/http"
	"sync"
)

// Settable allows you to wrap a handler and change it
type Settable struct {
	rwMtx sync.RWMutex
	next  http.Handler
}

// ServeHTTP makes Settable an http.Handler
func (h *Settable) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.rwMtx.RLock()
	defer h.rwMtx.RUnlock()
	h.next.ServeHTTP(w, r)
}

// Set sets a new handler, this replaces the old wrapped handler
func (h *Settable) Set(next http.Handler) {
	h.rwMtx.Lock()
	defer h.rwMtx.Unlock()
	h.next = next
}

// New returns a new settable handler that wraps next
func New(next http.Handler) *Settable {
	if next == nil {
		next = http.NotFoundHandler()
	}
	return &Settable{
		next: next,
	}
}
