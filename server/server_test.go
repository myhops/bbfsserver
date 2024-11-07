package server

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"
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

func TestCallback(t *testing.T) {
	getinfo := func() (*IndexPageInfo, error) {
		return nil, nil
	}
	responseBody := []byte("Hello there!!!")
	srv := New(
		slog.Default(),
		nil,
		[]*Version{
			{Name: "v1"},
			{Name: "v2"},
			{Name: "v3"},
			{Name: "v4"},
			{Name: "v5"},
			{Name: "v6"},
			{Name: "v7"},
		},
		nil,
		"",
		getinfo,
		time.Second,
		nil,
		WithControllerHandler(
			"test1",
			http.MethodGet,
			func(w http.ResponseWriter, r *http.Request) {
				w.Write(responseBody)
			}),
		WithControllerHandler(
			"test2",
			http.MethodGet,
			func(w http.ResponseWriter, r *http.Request) {
				w.Write(responseBody)
			}),
	)

	tsrv := httptest.NewServer(srv)
	defer tsrv.Close()

	f := func(name string) {
		path, err := url.JoinPath(tsrv.URL, ControllersPath, name)
		if err != nil {
			t.Fatalf("error: %s", err.Error())
		}

		resp, err := http.Get(path)
		if err != nil {
			t.Fatalf("error: %s", err.Error())
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("bad status code: %s", resp.Status)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("error: %s", err.Error())
		}
		if !slices.Equal(responseBody, body) {
			t.Errorf("expected %s, got %s", string(responseBody), string(body))
		}
	}
	f("test1")
	f("test2")

	t.Error()
}
