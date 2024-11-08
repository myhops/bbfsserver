package server

import (
	"net/url"
	"strings"
	"testing"
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

func TestIteratorSignature(t *testing.T) {
	s := &Server{
		versions: []*Version{
			{Name: "v1"},
			{Name: "v2"},
			{Name: "v3"},
			{Name: "v4"},
			{Name: "v5"},
			{Name: "v6"},
			{Name: "v7"},
		},
	}
	var i int
	for v := range s.Versions() {
		t.Logf("name %d: %s", i, v)
		i++
	}

	i = 0
	for v := range s.Versions() {
		t.Logf("name %d: %s", i, v)
		i++
		if !(i < 4) {
			break
		}
	}
}
