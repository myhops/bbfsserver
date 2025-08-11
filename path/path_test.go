package path

import (
	"net/url"
	"testing"
)

func TestPaths(t *testing.T) {
	cases := []struct {
		name  string
		paths []string
		want  string
	}{
		{
			name: "path 1",
			paths: []string{
				"/", "p1", "/",
			},
			want: "/p1/",
		},
		{
			name: "path 2",
			paths: []string{
				"/", "p1", "/",
			},
			want: "/p1/",
		},
		{
			name: "path 3",
			paths: []string{
				"/", "p1", "//p2",
			},
			want: "/p1/p2",
		},
		{
			name: "path 4",
			paths: []string{
				"http://host", "p1", "p2",
			},
			want: "http://host/p1/p2",
		},
		{
			name: "path 5",
			paths: []string{
				"/", "p1", "p2", "/",
			},
			want: "/p1/p2/",
		},
		{
			name: "path 5",
			paths: []string{
				"http://host", "p1", "p2", "/",
			},
			want: "http://host/p1/p2/",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := url.JoinPath(c.paths[0], c.paths[1:]...)
			if err != nil {
				t.Errorf("joinpath err: %s", err.Error())
				return
			}
			if res != c.want {
				t.Errorf("want: %s, got: %s", c.want, res)
			}
		})
	}
}
