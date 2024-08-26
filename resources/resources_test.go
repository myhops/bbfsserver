package resources

import (
	"io/fs"
	"testing"
)

func TestWalkDir(t *testing.T) {
	fs.WalkDir(StaticHtmlFS, ".", func(path string, d fs.DirEntry, err error) error {
		t.Logf("path %s, %s, %v", path, d.Name(), err)
		return nil
	})
}
