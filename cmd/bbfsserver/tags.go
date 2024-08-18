package main

import (
	"context"
	"log/slog"
	"net/url"
	"path/filepath"

	"github.com/myhops/bbfs"
	"github.com/myhops/bbfs/bbclient/server"
)

func getTags(cfg *bbfs.Config, logger *slog.Logger) ([]string, error) {
	u := url.URL{
		Scheme: "https",
		Host:   cfg.Host,
		Path:   filepath.Join(bbfs.ApiPath, bbfs.DefaultVersion),
	}

	// Find the valid tags
	client := server.Client{
		BaseURL:   u.String(),
		AccessKey: server.SecretString(cfg.AccessKey),
		Logger:    logger,
	}
	tags, err := client.GetTags(context.Background(), &server.GetTagsCommand{
		ProjectKey: cfg.ProjectKey,
		RepoSlug:   cfg.RepositorySlug,
	})
	if err != nil {
		return nil, err
	}
	return tags, nil
}
