package main

import (
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/myhops/bbfsserver/server"
)

type serverConfig struct {
	Logger          *slog.Logger
	AllFS           fs.FS
	Versions        []*server.Version
	WebFS           fs.FS
	IndexTemplate   string
	GetInfo         func() (*server.IndexPageInfo, error)
	TimeToLive      time.Duration
	CacheMiddleware func(next http.Handler) http.Handler
}
