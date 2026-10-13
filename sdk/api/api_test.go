package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	var selected string
	mux := http.NewServeMux()
	mux.HandleFunc("GET /version", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer s" {
			w.WriteHeader(401)
			w.Write([]byte(`{"message":"Unauthorized"}`))
			return
		}
		w.Write([]byte(`{"version":"v1"}`))
	})
	mux.HandleFunc("PUT /proxies/{group}", func(w http.ResponseWriter, r *http.Request) {
		var b struct{ Name string }
		json.NewDecoder(r.Body).Decode(&b)
		selected = r.PathValue("group") + "=" + b.Name
		w.WriteHeader(204)
	})
	mux.HandleFunc("GET /group/{name}/delay", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})
	mux.HandleFunc("GET /proxies/{name}/delay", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"delay":42}`))
	})
	mux.HandleFunc("GET /logs", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{\"type\":\"info\",\"payload\":\"a\"}\n{\"type\":\"info\",\"payload\":\"b\"}\n"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	ctx := context.Background()

	if _, err := New(srv.URL, "bad").Version(ctx); err == nil {
		t.Fatal("expected auth error")
	}
	c := New(srv.URL, "s")
	if v, err := c.Version(ctx); err != nil || v != "v1" {
		t.Fatalf("version = %q, %v", v, err)
	}
	if err := c.Select(ctx, "节点 选择", "HK 01"); err != nil || selected != "节点 选择=HK 01" {
		t.Fatalf("select = %q, %v", selected, err)
	}
	d, err := c.Delay(ctx, "HK 01", "http://x", time.Second)
	if err != nil || d["HK 01"] != 42 {
		t.Fatalf("delay = %v, %v", d, err)
	}
	var got []string
	if err := c.Logs(ctx, "info", func(e LogEntry) bool { got = append(got, e.Payload); return true }); err != nil || len(got) != 2 {
		t.Fatalf("logs = %v, %v", got, err)
	}
}

func TestConnectionsAndTraffic(t *testing.T) {
	var closed string
	mux := http.NewServeMux()
	mux.HandleFunc("GET /connections", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"uploadTotal":1,"downloadTotal":2,"connections":[{"id":"x","metadata":{"host":"a.com","destinationPort":"443"},"chains":["HK","Proxy"],"rule":"Match"}]}`))
	})
	mux.HandleFunc("DELETE /connections/{id}", func(w http.ResponseWriter, r *http.Request) {
		closed = r.PathValue("id")
		w.WriteHeader(204)
	})
	mux.HandleFunc("GET /traffic", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{\"up\":1,\"down\":2}\n{\"up\":3,\"down\":4}\n"))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	c := New(srv.URL, "")
	ctx := context.Background()

	cs, err := c.Connections(ctx)
	if err != nil || len(cs.Connections) != 1 || cs.Connections[0].Target() != "a.com:443" {
		t.Fatalf("connections = %+v, %v", cs, err)
	}
	if err := c.CloseConnection(ctx, "x"); err != nil || closed != "x" {
		t.Fatalf("close = %q, %v", closed, err)
	}
	var last Traffic
	if err := c.Traffic(ctx, func(tr Traffic) bool { last = tr; return true }); err != nil || last.Down != 4 {
		t.Fatalf("traffic = %+v, %v", last, err)
	}
}
