package server

import (
	"net/url"
	"strings"
)

func getIndexPageInfo(
	repoURL string,
	title string,
	projectKey string,
	repositorySlug string,
	tags []string,
) func() (*IndexPageInfo, error) {
	url := &url.URL{
		Path: "/versions",
	}
	var versions []struct {
		Name string
		Path string
	}
	for _, tag := range tags {
		parts := strings.Split(tag, "/")
		module := ""
		if len(parts) == 2 {
			module = parts[0]
		}
		v := struct {
			Name string
			Path string
		}{
			Name: tag,
			Path: url.JoinPath(tag, module, "/").String(),
		}
		versions = append(versions, v)
	}

	return func() (*IndexPageInfo, error) {
		res := &IndexPageInfo{
			BitbucketURL:   repoURL,
			Title:          title,
			ProjectKey:     projectKey,
			RepositorySlug: repositorySlug,
			Versions:       versions,
		}
		return res, nil
	}
}
