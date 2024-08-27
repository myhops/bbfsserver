package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"path/filepath"

	"github.com/myhops/bbfsserver/server"

	"github.com/myhops/bbfs"
	bbfsserver "github.com/myhops/bbfs/bbclient/server"
)

func getTagsNil(cfg *bbfs.Config, logger *slog.Logger) []string {
	res, err := getTags(cfg, logger)
	if err != nil {
		return nil
	} 
	return res
}

func getTags(cfg *bbfs.Config, logger *slog.Logger) ([]string, error) {
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
	tags, err := client.GetTags(context.Background(), &bbfsserver.GetTagsCommand{
		ProjectKey: cfg.ProjectKey,
		RepoSlug:   cfg.RepositorySlug,
	})
	if err != nil {
		return nil, err
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

func getVersionsFromTags(cfg *bbfs.Config, logger *slog.Logger, tags []string) ([]*server.Version, error) {
	var c *bbfs.Config
	*c = *cfg
	cfg = nil // make sure we do not use it

	res := make([]*server.Version, 0, len(tags))
	for _, tag := range tags {
		// Create a fs for the tag.
		c.At = tag
		res = append(res, &server.Version{
			Name: tag,
			Dir:  bbfs.NewFS(c),
		})
	}
	return res, nil
}

func getVersionNames(versions []*server.Version) []string {
	res := make([]string, 0, len(versions))
	for _, v := range versions {
		res = append(res, v.Name)
	}
	return res
}

