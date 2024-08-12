package main

import (
	"os"
	"testing"
	"time"
)

func TestGetOptionsFromEnv(t *testing.T) {


	os.Setenv("PORT", "10000")
	os.Setenv("BBFSSRV_HOST", "BBHOST.example.com")
	os.Setenv("BBFSSRV_PROJECT_KEY", "projectKey")
	os.Setenv("BBFSSRV_REPOSITORY_SLUG", "repoSlug")
	os.Setenv("BBFSSRV_ACCESS_KEY", "accessKey")
	os.Setenv("BBFSSRV_LOG_FORMAT","json")
	os.Setenv("BBFSSRV_TAG_POLL_INTERVAL", "1ns")

	opts := &options{}
	opts.fromEnv()

	if opts.tagsPollInterval != time.Second {
		t.Errorf("want %v, got %v", time.Minute, opts.tagsPollInterval)
	}

}
