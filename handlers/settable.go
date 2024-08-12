package handlers

import (
	"net/http"
	"sync"
)

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

func (h *Settable) Set(next http.Handler) {
	h.rwMtx.Lock()
	defer h.rwMtx.Unlock()
	h.next = next
}

func NewSettable(next http.Handler) *Settable {
	return &Settable{
		next: next,
	}
}
