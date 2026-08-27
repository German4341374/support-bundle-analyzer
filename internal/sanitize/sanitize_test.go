package sanitize

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWorkspaceRedactsTextAndExcludesBinary(t *testing.T) {
	t.Parallel()
	source := t.TempDir()
	if err := os.MkdirAll(filepath.Join(source, "artifacts"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "artifacts", "app.log"), []byte("email=alex@example.test password=do-not-share"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "artifacts", "dump.bin"), []byte{'a', 0, 'b'}, 0o600); err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(t.TempDir(), "sanitized")
	manifest, err := Workspace(source, destination, "strict")
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(destination, "artifacts", "app.log"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "alex@example.test") || strings.Contains(string(content), "do-not-share") {
		t.Fatalf("sensitive values remain: %q", content)
	}
	if _, err := os.Stat(filepath.Join(destination, "artifacts", "dump.bin")); !os.IsNotExist(err) {
		t.Fatal("binary artifact was copied")
	}
	if len(manifest.Files) != 2 || manifest.Summary["email"] != 1 {
		t.Fatalf("unexpected manifest: %+v", manifest)
	}
}

func TestWorkspaceRefusesOverwrite(t *testing.T) {
	t.Parallel()
	source := t.TempDir()
	if err := os.Mkdir(filepath.Join(source, "artifacts"), 0o700); err != nil {
		t.Fatal(err)
	}
	destination := t.TempDir()
	if _, err := Workspace(source, destination, "standard"); err == nil {
		t.Fatal("expected overwrite refusal")
	}
}

func TestWorkspaceRejectsSymlinkArtifacts(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks normally requires elevated Windows privileges")
	}
	source := t.TempDir()
	if err := os.Mkdir(filepath.Join(source, "artifacts"), 0o700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(target, []byte("outside secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(source, "artifacts", "linked.log")); err != nil {
		t.Fatal(err)
	}
	_, err := Workspace(source, filepath.Join(t.TempDir(), "sanitized"), "strict")
	if err == nil || !strings.Contains(err.Error(), "non-regular") {
		t.Fatalf("expected non-regular artifact rejection, got %v", err)
	}
}
