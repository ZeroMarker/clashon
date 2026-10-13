// Package update checks GitHub Releases for a newer ClashOn.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Repo is the GitHub repository releases are published to.
const Repo = "ZeroMarker/clashon"

// Release is the subset of the GitHub release we use.
type Release struct {
	Version string `json:"version"` // without the leading v
	URL     string `json:"url"`     // release page
}

// Latest returns the newest non-prerelease release.
func Latest(ctx context.Context, client *http.Client, apiBase string) (*Release, error) {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBase+"/repos/"+Repo+"/releases/latest", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // no releases yet
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github: HTTP %d", resp.StatusCode)
	}
	var r struct {
		Tag string `json:"tag_name"`
		URL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &Release{Version: strings.TrimPrefix(r.Tag, "v"), URL: r.URL}, nil
}

// Newer reports whether version a is newer than b. Both are "X.Y.Z" with an
// optional "-pre" suffix; dev builds (non-numeric) never count as newer and
// are always considered older than a release.
func Newer(a, b string) bool {
	pa, oka := parse(a)
	pb, okb := parse(b)
	if !oka {
		return false
	}
	if !okb {
		return true
	}
	for i := range pa.n {
		if pa.n[i] != pb.n[i] {
			return pa.n[i] > pb.n[i]
		}
	}
	// 1.0.0 > 1.0.0-rc1
	return pa.pre == "" && pb.pre != ""
}

type ver struct {
	n   [3]int
	pre string
}

func parse(s string) (ver, bool) {
	s = strings.TrimPrefix(s, "v")
	core, pre, _ := strings.Cut(s, "-")
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return ver{}, false
	}
	var v ver
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return ver{}, false
		}
		v.n[i] = n
	}
	v.pre = pre
	return v, true
}
