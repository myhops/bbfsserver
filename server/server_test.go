package server

// func getIndexPageInfo(
// 	repoURL string,
// 	title string,
// 	projectKey string,
// 	repositorySlug string,
// 	tags []string,
// ) func() (*IndexPageInfo, error) {
// 	url := &url.URL{
// 		Path: "/versions",
// 	}
// 	var versions []struct {
// 		Name string
// 		Path string
// 	}
// 	for _, tag := range tags {
// 		parts := strings.Split(tag, "/")
// 		module := ""
// 		if len(parts) == 2 {
// 			module = parts[0]
// 		}
// 		v := struct {
// 			Name string
// 			Path string
// 		}{
// 			Name: tag,
// 			Path: url.JoinPath(tag, module, "/").String(),
// 		}
// 		versions = append(versions, v)
// 	}

// 	return func() (*IndexPageInfo, error) {
// 		res := &IndexPageInfo{
// 			Title:          title,
// 			ProjectKey:     projectKey,
// 			RepositorySlug: repositorySlug,
// 			Versions:       versions,
// 		}
// 		return res, nil
// 	}
// }

// func TestIndexPage(t *testing.T) {
// 	out := &bytes.Buffer{}
// 	logger := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{}))

// 	tags := []string{"tag1", "tag2"}
// 	cfg := &bbfs.Config{}
// 	getinfo := getIndexPageInfo("Title", "Project 1", "Repo 1", []string{"tag1"})
// 	srv := New(cfg, logger, tags, staticHtmlFS, indexHtmlTemplate, getinfo)
// 	h := srv.indexPageHandler(indexHtmlTemplate, getinfo)

// 	r := httptest.NewRequest(http.MethodGet, "/", nil)
// 	w := httptest.NewRecorder()
// 	h.ServeHTTP(w, r)
// 	body, err := io.ReadAll(w.Result().Body)
// 	if err != nil {
// 		t.Errorf("error reading body: %s", err.Error())
// 	}
// 	bodys := string(body)
// 	_ = bodys
// 	t.Logf("status: %s", w.Result().Status)
// }

// func TestIndexPageWithServer(t *testing.T) {
// 	out := &bytes.Buffer{}
// 	logger := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{}))

// 	tags := []string{"tag1", "tag2"}
// 	cfg := &bbfs.Config{}
// 	getinfo := getIndexPageInfo("Title", "Project 1", "Repo 1", []string{"tag1"})
// 	h := New(cfg, logger, tags, staticHtmlFS, indexHtmlTemplate, getinfo)
// 	srv := httptest.NewServer(h)
// 	defer srv.Close()
// 	u := srv.URL
// 	_ = u

// 	r, err := http.Get(srv.URL)
// 	if err != nil {
// 		t.Errorf("error getting page: %s", err.Error())
// 	}
// 	defer r.Body.Close()

// 	body, err := io.ReadAll(r.Body)
// 	if err != nil {
// 		t.Errorf("error reading body: %s", err.Error())
// 	}
// 	bodys := string(body)
// 	_ = bodys
// 	t.Logf("status: %s", r.Status)
// }
