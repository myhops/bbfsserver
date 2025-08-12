package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/myhops/bbfsserver/server"

	"github.com/myhops/bbfs"
	bbfsserver "github.com/myhops/bbfs/bbclient/server"
)

// getTags returns all tags (max 1000)
func getTags(cfg *bbfs.Config, logger *slog.Logger) ([]string, error) {
	logger = logger.With(slog.String("method", "getTags"))
	u := url.URL{
		Scheme: "https",
		Host:   cfg.Host,
		Path:   filepath.Join(bbfs.ApiPath, bbfs.DefaultVersion),
	}

	// Find the valid tags
	client := bbfsserver.Client{
		BaseURL:   u.String(),
		AccessKey: bbfsserver.SecretString(cfg.AccessKey),
		Logger:    logger,
	}
	resp, err := client.GetTags(context.Background(), &bbfsserver.GetTagsCommand{
		ProjectKey: cfg.ProjectKey,
		RepoSlug:   cfg.RepositorySlug,
		Limit:      1000,
	})
	if err != nil {
		return nil, err
	}
	tags := make([]string, 0, len(resp.Tags))
	for _, tag := range resp.Tags {
		if !strings.Contains(tag.Name, "/") {
			logger.Debug("skipped tag", slog.String("name", tag.Name), slog.String("type", tag.Type))
			continue
		}
		logger.Debug("adding tag", slog.String("name", tag.Name))
		tags = append(tags, tag.Name)
	}
	return tags, nil
}

func getVersions(cfg *bbfs.Config, logger *slog.Logger) ([]*server.Version, error) {
	tags, err := getTags(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("error getting versions: %s", err)
	}
	return getVersionsFromTags(cfg, logger, tags)
}

func getVersionsFromTags(cfg *bbfs.Config, _ *slog.Logger, tags []string) ([]*server.Version, error) {
	c := *cfg
	cfg = nil // make sure we do not use it

	res := make([]*server.Version, 0, len(tags))
	for _, tag := range tags {
		// Create a fs for the tag.
		c.At = tag
		res = append(res, &server.Version{
			Name: tag,
			Dir:  bbfs.NewFS(&c),
		})
	}
	return res, nil
}

