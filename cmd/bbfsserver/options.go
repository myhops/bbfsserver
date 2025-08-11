package main

import (
	_ "embed"
	"time"
)

//go:embed usage.txt
var usageText string

type options struct {
	host                  string
	logFormat             string
	listenAddress         string
	projectKey            string
	repositorySlug        string
	accessKey             string
	changePollingInterval time.Duration
	dryRun                string
	repoURL               string
	title                 string
}

func defaultOptions() *options {
	return &options{
		logFormat:             "json",
		listenAddress:         ":8080",
		changePollingInterval: 5 * time.Minute,
		title: "BBFS Server Rocks (use env var BBFSSRV_TITLE to set the title",
	}
}

func setIfSet(v string, val *string) {
	if v != "" {
		*val = v
	}
}

// setIfSetDuration sets duration from v if v is a valid duration.
// The minumum value is 1 second.
func setIfSetDuration(v string, dp *time.Duration) {
	if v == "" {
		return
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return
	}
	if d < time.Second {
		d = time.Second
	}
	*dp = d
}

func (o *options) fromEnv(getenv func(string) string) {
	setIfSet(getenv("PORT"), &o.listenAddress)
	setIfSet(getenv("BBFSSRV_LISTEN_ADDRESS"), &o.listenAddress)
	setIfSet(getenv("BBFSSRV_HOST"), &o.host)
	setIfSet(getenv("BBFSSRV_PROJECT_KEY"), &o.projectKey)
	setIfSet(getenv("BBFSSRV_REPOSITORY_SLUG"), &o.repositorySlug)
	setIfSet(getenv("BBFSSRV_ACCESS_KEY"), &o.accessKey)
	setIfSet(getenv("BBFSSRV_LOG_FORMAT"), &o.logFormat)
	setIfSet(getenv("BBFSSRV_DRY_RUN"), &o.dryRun)
	setIfSet(getenv("BBFSSRV_REPO_URL"), &o.repoURL)
	setIfSet(getenv("BBFSSRV_TITLE"), &o.title)
	setIfSetDuration(getenv("BBFSSRV_CHANGE_POLLING_INTERVAL"), &o.changePollingInterval)

	// fix listen address if needed.
	if o.listenAddress[0] != ':' {
		o.listenAddress = ":" + o.listenAddress
	}
}
