package handlers

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

// NewSettable returns a new settable handler that wraps next
func NewSettable(next http.Handler) *Settable {
	return &Settable{
		next: next,
	}
}
