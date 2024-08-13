package main

import (
	"testing"
	"time"
)

func testGetOptionsFromEnvGetenv(key string) string {
	switch key {
	case "PORT":
		return "10000"
	case "BBFSSRV_HOST":
		return "BBHOST.example.com"
	case "BBFSSRV_PROJECT_KEY":
		return "projectKey"
	case "BBFSSRV_REPOSITORY_SLUG":
		return "repoSlug"
	case "BBFSSRV_ACCESS_KEY":
		return "accessKey"
	case "BBFSSRV_LOG_FORMAT":
		return "json"
	case "BBFSSRV_TAG_POLL_INTERVAL":
		return "1ns"
	default:
		return ""
	}
}

func TestGetOptionsFromEnv(t *testing.T) {
	opts := &options{}
	opts.fromEnv(testGetOptionsFromEnvGetenv)

	if opts.tagsPollInterval != time.Second {
		t.Errorf("want %v, got %v", time.Minute, opts.tagsPollInterval)
	}
}
