package profile

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ZeroMarker/clashon/sdk/fsutil"
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

// The GUI polls the index (via `clashon sub ls`/`status`) while other
// commands rewrite it. On Windows, replacing a file another process has open
// fails unless the writer retries.
func TestConcurrentReadWrite(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("proxies:\n  - {name: a, type: direct}\n"))
	}))
	defer srv.Close()
	dir := t.TempDir()
	s := &Store{Dir: filepath.Join(dir, "p"), Index: filepath.Join(dir, "i.json")}
	for _, n := range []string{"a", "b"} {
		if _, err := s.Add(context.Background(), n, srv.URL); err != nil {
			t.Fatal(err)
		}
	}

	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				return
			default:
			}
			if f, err := fsutil.Open(s.Index); err == nil {
				b := make([]byte, 64)
				f.Read(b)
				time.Sleep(time.Millisecond) // hold it open like a slow reader
				f.Close()
			}
		}
	}()
	defer func() { close(stop); <-done }()

	for i := 0; i < 200; i++ {
		if err := s.Use([]string{"a", "b"}[i%2]); err != nil {
			t.Fatalf("write %d while a reader has the index open: %v", i, err)
		}
	}
}
