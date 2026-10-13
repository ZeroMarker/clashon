package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewer(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want bool
	}{
		{"1.2.0", "1.1.9", true},
		{"1.1.9", "1.2.0", false},
		{"1.0.0", "1.0.0", false},
		{"1.0.0", "1.0.0-rc1", true},
		{"1.0.0-rc1", "1.0.0", false},
		{"0.1.0", "dev", true},
		{"0.1.0", "0.0.0-dev.abc1234", true},
		{"dev", "0.1.0", false},
	} {
		if got := Newer(tc.a, tc.b); got != tc.want {
			t.Errorf("Newer(%q, %q) = %v", tc.a, tc.b, got)
		}
	}
}

func TestLatest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/"+Repo+"/releases/latest" {
			w.WriteHeader(404)
			return
		}
		w.Write([]byte(`{"tag_name":"v1.2.3","html_url":"https://github.com/x/releases/v1.2.3"}`))
	}))
	defer srv.Close()
	r, err := Latest(context.Background(), nil, srv.URL)
	if err != nil || r.Version != "1.2.3" {
		t.Fatalf("got %+v, %v", r, err)
	}
}
