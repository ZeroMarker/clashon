// Package subscription downloads Clash subscriptions.
package subscription

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// UserAgent makes providers return Clash.Meta YAML instead of base64 links.
const UserAgent = "clash.meta"

// UserInfo is the subscription-userinfo header.
type UserInfo struct {
	Upload   int64     `json:"upload"`
	Download int64     `json:"download"`
	Total    int64     `json:"total"`
	Expire   time.Time `json:"expire,omitempty"`
}

// Result is a downloaded subscription.
type Result struct {
	Body     []byte
	Filename string // from Content-Disposition, may be empty
	UserInfo *UserInfo
	Interval time.Duration // from profile-update-interval, may be 0
}

// Fetch downloads url and checks that it is a Clash config with proxies.
func Fetch(ctx context.Context, client *http.Client, url string) (*Result, error) {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("subscription: HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, err
	}
	if err := Validate(body); err != nil {
		return nil, err
	}
	r := &Result{Body: body, UserInfo: ParseUserInfo(resp.Header.Get("Subscription-Userinfo"))}
	if _, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition")); err == nil {
		r.Filename = params["filename"]
	}
	if h, err := strconv.Atoi(resp.Header.Get("Profile-Update-Interval")); err == nil && h > 0 {
		r.Interval = time.Duration(h) * time.Hour
	}
	return r, nil
}

// Validate reports whether body is a Clash YAML config with at least one proxy
// or proxy provider.
func Validate(body []byte) error {
	var doc struct {
		Proxies   []map[string]any `yaml:"proxies"`
		Providers map[string]any   `yaml:"proxy-providers"`
	}
	if err := yaml.Unmarshal(body, &doc); err != nil {
		return fmt.Errorf("subscription: not a Clash YAML config: %w", err)
	}
	if len(doc.Proxies) == 0 && len(doc.Providers) == 0 {
		return fmt.Errorf("subscription: no proxies or proxy-providers")
	}
	return nil
}

// ParseUserInfo parses "upload=1; download=2; total=3; expire=1700000000".
func ParseUserInfo(h string) *UserInfo {
	if h == "" {
		return nil
	}
	info := &UserInfo{}
	for _, part := range strings.Split(h, ";") {
		k, v, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil {
			continue
		}
		switch strings.ToLower(k) {
		case "upload":
			info.Upload = n
		case "download":
			info.Download = n
		case "total":
			info.Total = n
		case "expire":
			if n > 0 {
				info.Expire = time.Unix(n, 0)
			}
		}
	}
	return info
}
