package main

import (
	"slices"
	"time"
	_ "embed"
)

//go:embed usage.txt
var usageText string

type options struct {
	host             string
	logFormat        string
	listenAddress    string
	projectKey       string
	repositorySlug   string
	accessKey        string
	tagsPollInterval time.Duration
	dryRun           string
	repoURL          string
}

func defaultOptions() *options {
	return &options{
		logFormat:     "json",
		listenAddress: ":8080",
	}
}

func setIfSet(v string, val *string) {
	if v != "" {
		*val = v
	}
}

// TODO: Cleanup
func compareTags(t1, t2 []string) int {
	slices.Sort(t1)
	slices.Sort(t2)
	return slices.Compare(t1, t2)
}

func getPollInterval(interval string) time.Duration {
	res := newTagPollingInterval
	if interval == "" {
		return res
	}
	i, err := time.ParseDuration(interval)
	if err != nil {
		return res
	}
	if i < time.Second {
		return time.Second
	}
	res = i
	return res
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

	o.tagsPollInterval = getPollInterval(getenv("BBFSSRV_TAG_POLL_INTERVAL"))

	// fix listen address if needed.
	if o.listenAddress[0] != ':' {
		o.listenAddress = ":" + o.listenAddress
	}
}
