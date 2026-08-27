package archiveutil

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/German4341374/support-bundle-analyzer/internal/model"
)

func TestExtractZIP(t *testing.T) {
	t.Parallel()
	archive := createZIP(t, []zipEntry{{name: "logs/api.log", body: "2026-01-01 ERROR connection refused"}, {name: "данные/info.txt", body: "ok"}})
	destination := filepath.Join(t.TempDir(), "out")
	result, err := Extract(archive, destination, testLimits())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 2 || result.InputSHA256 == "" {
		t.Fatalf("unexpected extraction result: %+v", result)
	}
	content, err := os.ReadFile(filepath.Join(destination, "logs", "api.log"))
	if err != nil || !bytes.Contains(content, []byte("connection refused")) {
		t.Fatalf("extracted content missing: %v", err)
	}
}

func TestExtractRejectsUnsafeZIPEntries(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		entries []zipEntry
	}{
		{name: "parent traversal", entries: []zipEntry{{name: "../../escape.txt", body: "no"}}},
		{name: "unix absolute", entries: []zipEntry{{name: "/etc/passwd", body: "no"}}},
		{name: "windows absolute", entries: []zipEntry{{name: `C:\Windows\win.ini`, body: "no"}}},
		{name: "duplicate normalized", entries: []zipEntry{{name: "logs/a.log", body: "one"}, {name: "logs/./a.log", body: "two"}}},
		{name: "symlink", entries: []zipEntry{{name: "escape", body: "target", symlink: true}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			archive := createZIP(t, test.entries)
			if _, err := Extract(archive, filepath.Join(t.TempDir(), "out"), testLimits()); err == nil {
				t.Fatal("unsafe archive was accepted")
			}
		})
	}
}

func TestExtractRejectsCompressionRatio(t *testing.T) {
	t.Parallel()
	archive := createZIP(t, []zipEntry{{name: "large.log", body: strings.Repeat("0", 32<<10)}})
	limits := testLimits()
	limits.MaxCompressionRatio = 2
	if _, err := Extract(archive, filepath.Join(t.TempDir(), "out"), limits); err == nil {
		t.Fatal("high compression ratio archive was accepted")
	}
}

func TestExtractRejectsTARSymlink(t *testing.T) {
	t.Parallel()
	name := filepath.Join(t.TempDir(), "unsafe.tar")
	f, err := os.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	w := tar.NewWriter(f)
	if err := w.WriteHeader(&tar.Header{Name: "escape", Typeflag: tar.TypeSymlink, Linkname: "../../outside"}); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := Extract(name, filepath.Join(t.TempDir(), "out"), testLimits()); err == nil {
		t.Fatal("TAR symlink was accepted")
	}
}

type zipEntry struct {
	name    string
	body    string
	symlink bool
}

func createZIP(t *testing.T, entries []zipEntry) string {
	t.Helper()
	name := filepath.Join(t.TempDir(), "fixture.zip")
	f, err := os.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.name, Method: zip.Deflate}
		if entry.symlink {
			header.SetMode(os.ModeSymlink | 0o777)
		}
		part, err := w.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write([]byte(entry.body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	return name
}

func testLimits() model.Limits {
	limits := model.DefaultLimits()
	limits.MaxFiles = 20
	limits.MaxSingleFileBytes = 1 << 20
	limits.MaxTotalBytes = 4 << 20
	return limits
}
