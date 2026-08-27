package synthetic

import (
	"archive/zip"
	"path/filepath"
	"testing"
)

func TestWriteBundleIsReadable(t *testing.T) {
	t.Parallel()
	name := filepath.Join(t.TempDir(), "demo.zip")
	if err := WriteBundle(name); err != nil {
		t.Fatal(err)
	}
	reader, err := zip.OpenReader(name)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	if len(reader.File) != 4 {
		t.Fatalf("got %d files", len(reader.File))
	}
}
