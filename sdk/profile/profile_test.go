package profile

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestStore(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("proxies:\n  - {name: a, type: direct}\n"))
	}))
	defer srv.Close()
	dir := t.TempDir()
	s := &Store{Dir: filepath.Join(dir, "p"), Index: filepath.Join(dir, "i.json")}
	ctx := context.Background()

	if _, err := s.Add(ctx, "one", srv.URL); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Add(ctx, "two", srv.URL); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Add(ctx, "one", srv.URL); err == nil {
		t.Fatal("duplicate add should fail")
	}
	if _, err := s.Add(ctx, "../x", srv.URL); err == nil {
		t.Fatal("path traversal name should fail")
	}
	cur, err := s.Current()
	if err != nil || cur.Name != "one" {
		t.Fatalf("current = %v, %v", cur, err)
	}
	if err := s.Use("two"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(ctx, "two"); err != nil {
		t.Fatal(err)
	}
	if err := s.Remove("two"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(s.File("two")); !os.IsNotExist(err) {
		t.Fatal("file not removed")
	}
	if _, err := s.Current(); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	ps, _, _ := s.List()
	if len(ps) != 1 {
		t.Fatalf("list = %d", len(ps))
	}
}
