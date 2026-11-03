package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

// A file held open by a reader must still be replaceable (Windows).
func TestReplaceWhileOpen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "f.json")
	if err := WriteFile(path, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	f, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := WriteFile(path, []byte("new"), 0o600); err != nil {
		t.Fatalf("replace while open: %v", err)
	}
	b, err := ReadFile(path)
	if err != nil || string(b) != "new" {
		t.Fatalf("read back %q, %v", b, err)
	}
	if _, err := ReadFile(filepath.Join(t.TempDir(), "missing")); !os.IsNotExist(err) {
		t.Fatalf("missing file: %v", err)
	}
}
