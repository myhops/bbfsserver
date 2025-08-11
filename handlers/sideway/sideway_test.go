package sideway

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSideway(t *testing.T) {
	var handlerCalled bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	var nextCalled bool
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	srv := New(next, slog.Default())
	srv.HandleFunc("/api/controllers", handler)

	r := httptest.NewRequest(http.MethodGet, "/api/controllers", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, r)

	if nextCalled {
		t.Logf("nextCalled")
	}
	if handlerCalled {
		t.Logf("handlerCalled")
	}

	handlerCalled = false
	nextCalled = false
	r = httptest.NewRequest(http.MethodGet, "/api/controllers/not-present", nil)
	w = httptest.NewRecorder()

	srv.ServeHTTP(w, r)

	if nextCalled {
		t.Logf("nextCalled")
	}
	if handlerCalled {
		t.Logf("handlerCalled")
	}

}
