package subscription

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

const sample = `proxies:
  - {name: a, type: ss, server: 1.1.1.1, port: 443, cipher: aes-128-gcm, password: x}
`

func TestParseUserInfo(t *testing.T) {
	u := ParseUserInfo("upload=1; download=2; total=1024; expire=1700000000")
	if u.Upload != 1 || u.Download != 2 || u.Total != 1024 || u.Expire.Unix() != 1700000000 {
		t.Fatalf("got %+v", u)
	}
	if ParseUserInfo("") != nil {
		t.Fatal("empty header should be nil")
	}
}

func TestValidate(t *testing.T) {
	if err := Validate([]byte(sample)); err != nil {
		t.Fatal(err)
	}
	if Validate([]byte("c3M6Ly8=")) == nil {
		t.Fatal("base64 link list should be rejected")
	}
}

func TestFetch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.UserAgent() != UserAgent {
			t.Errorf("UA = %q", r.UserAgent())
		}
		w.Header().Set("Subscription-Userinfo", "upload=1; download=1; total=10")
		w.Header().Set("Profile-Update-Interval", "24")
		w.Header().Set("Content-Disposition", `attachment; filename="air.yaml"`)
		w.Write([]byte(sample))
	}))
	defer srv.Close()
	r, err := Fetch(context.Background(), nil, srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if r.Filename != "air.yaml" || r.Interval != 24*time.Hour || r.UserInfo.Total != 10 {
		t.Fatalf("got %+v", r)
	}
}

// TestFetchLive hits the real subscription in CLASHON_TEST_SUB_URL.
func TestFetchLive(t *testing.T) {
	url := os.Getenv("CLASHON_TEST_SUB_URL")
	if url == "" {
		t.Skip("CLASHON_TEST_SUB_URL not set")
	}
	r, err := Fetch(context.Background(), nil, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("fetched %d bytes, userinfo=%+v", len(r.Body), r.UserInfo)
}
