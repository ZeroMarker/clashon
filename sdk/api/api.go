// Package api is a client for the mihomo external-controller REST API.
package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client talks to one controller, e.g. "127.0.0.1:9090".
type Client struct {
	Addr   string
	Secret string
	HTTP   *http.Client
}

// New returns a client with a short default timeout.
func New(addr, secret string) *Client {
	return &Client{Addr: addr, Secret: secret, HTTP: &http.Client{Timeout: 10 * time.Second}}
}

type Proxy struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Now     string   `json:"now,omitempty"`
	All     []string `json:"all,omitempty"`
	History []struct {
		Delay int `json:"delay"`
	} `json:"history"`
}

// IsGroup reports whether the proxy is a selectable group.
func (p *Proxy) IsGroup() bool { return len(p.All) > 0 }

// LastDelay returns the most recent delay in ms, or 0 if unknown/failed.
func (p *Proxy) LastDelay() int {
	if len(p.History) == 0 {
		return 0
	}
	return p.History[len(p.History)-1].Delay
}

type Rule struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
	Proxy   string `json:"proxy"`
}

type Configs struct {
	MixedPort int    `json:"mixed-port"`
	Mode      string `json:"mode"`
	LogLevel  string `json:"log-level"`
	AllowLAN  bool   `json:"allow-lan"`
}

type LogEntry struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

func (c *Client) Version(ctx context.Context) (string, error) {
	var v struct {
		Version string `json:"version"`
	}
	return v.Version, c.do(ctx, http.MethodGet, "/version", nil, &v)
}

// Proxies returns all proxies and groups keyed by name.
func (c *Client) Proxies(ctx context.Context) (map[string]*Proxy, error) {
	var v struct {
		Proxies map[string]*Proxy `json:"proxies"`
	}
	return v.Proxies, c.do(ctx, http.MethodGet, "/proxies", nil, &v)
}

// Select chooses proxy in group.
func (c *Client) Select(ctx context.Context, group, proxy string) error {
	return c.do(ctx, http.MethodPut, "/proxies/"+url.PathEscape(group), map[string]string{"name": proxy}, nil)
}

// Delay tests a proxy or group and returns ms; groups return per-member delays.
func (c *Client) Delay(ctx context.Context, name, testURL string, timeout time.Duration) (map[string]int, error) {
	q := url.Values{"url": {testURL}, "timeout": {strconv.Itoa(int(timeout.Milliseconds()))}}
	esc := url.PathEscape(name)
	// Try group endpoint first; fall back to single proxy.
	res := map[string]int{}
	err := c.do(ctx, http.MethodGet, "/group/"+esc+"/delay?"+q.Encode(), nil, &res)
	if err == nil {
		return res, nil
	}
	var single struct {
		Delay int `json:"delay"`
	}
	if err := c.do(ctx, http.MethodGet, "/proxies/"+esc+"/delay?"+q.Encode(), nil, &single); err != nil {
		return nil, err
	}
	return map[string]int{name: single.Delay}, nil
}

func (c *Client) Rules(ctx context.Context) ([]Rule, error) {
	var v struct {
		Rules []Rule `json:"rules"`
	}
	return v.Rules, c.do(ctx, http.MethodGet, "/rules", nil, &v)
}

func (c *Client) Configs(ctx context.Context) (*Configs, error) {
	v := &Configs{}
	return v, c.do(ctx, http.MethodGet, "/configs", nil, v)
}

// PatchConfigs updates runtime settings, e.g. {"mode": "global"}.
func (c *Client) PatchConfigs(ctx context.Context, patch map[string]any) error {
	return c.do(ctx, http.MethodPatch, "/configs", patch, nil)
}

// Logs streams log entries at level until ctx is done or fn returns false.
func (c *Client) Logs(ctx context.Context, level string, fn func(LogEntry) bool) error {
	return stream(c, ctx, "/logs?level="+url.QueryEscape(level), fn)
}

// Traffic streams per-second throughput until ctx is done or fn returns false.
func (c *Client) Traffic(ctx context.Context, fn func(Traffic) bool) error {
	return stream(c, ctx, "/traffic", fn)
}

type Traffic struct {
	Up   int64 `json:"up"`
	Down int64 `json:"down"`
}

type Connection struct {
	ID       string `json:"id"`
	Metadata struct {
		Network         string `json:"network"`
		Type            string `json:"type"`
		SourceIP        string `json:"sourceIP"`
		DestinationIP   string `json:"destinationIP"`
		DestinationPort string `json:"destinationPort"`
		Host            string `json:"host"`
		Process         string `json:"process"`
	} `json:"metadata"`
	Upload      int64     `json:"upload"`
	Download    int64     `json:"download"`
	Start       time.Time `json:"start"`
	Chains      []string  `json:"chains"`
	Rule        string    `json:"rule"`
	RulePayload string    `json:"rulePayload"`
}

// Target returns host:port, falling back to the destination IP.
func (c *Connection) Target() string {
	h := c.Metadata.Host
	if h == "" {
		h = c.Metadata.DestinationIP
	}
	return h + ":" + c.Metadata.DestinationPort
}

type Connections struct {
	DownloadTotal int64        `json:"downloadTotal"`
	UploadTotal   int64        `json:"uploadTotal"`
	Connections   []Connection `json:"connections"`
}

func (c *Client) Connections(ctx context.Context) (*Connections, error) {
	v := &Connections{}
	return v, c.do(ctx, http.MethodGet, "/connections", nil, v)
}

// CloseConnection closes one connection, or all if id is empty.
func (c *Client) CloseConnection(ctx context.Context, id string) error {
	p := "/connections"
	if id != "" {
		p += "/" + url.PathEscape(id)
	}
	return c.do(ctx, http.MethodDelete, p, nil, nil)
}

// stream reads newline-delimited JSON from a streaming endpoint.
func stream[T any](c *Client, ctx context.Context, path string, fn func(T) bool) error {
	req, err := c.request(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	client := &http.Client{Transport: c.httpClient().Transport} // no timeout for streams
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return err
	}
	defer resp.Body.Close()
	if err := checkStatus(resp); err != nil {
		return err
	}
	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		var v T
		if json.Unmarshal(sc.Bytes(), &v) != nil {
			continue
		}
		if !fn(v) {
			return nil
		}
	}
	if ctx.Err() != nil {
		return nil
	}
	return sc.Err()
}

func (c *Client) httpClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return http.DefaultClient
}

func (c *Client) request(ctx context.Context, method, path string, body any) (*http.Request, error) {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(b)
	}
	base := c.Addr
	if !strings.Contains(base, "://") {
		base = "http://" + base
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(base, "/")+path, r)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Secret != "" {
		req.Header.Set("Authorization", "Bearer "+c.Secret)
	}
	return req, nil
}

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	req, err := c.request(ctx, method, path, body)
	if err != nil {
		return err
	}
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := checkStatus(resp); err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func checkStatus(resp *http.Response) error {
	if resp.StatusCode < 300 {
		return nil
	}
	var e struct {
		Message string `json:"message"`
	}
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if json.Unmarshal(b, &e) == nil && e.Message != "" {
		return fmt.Errorf("controller: %s (HTTP %d)", e.Message, resp.StatusCode)
	}
	return fmt.Errorf("controller: HTTP %d", resp.StatusCode)
}

// Reload makes the core re-read the config file at path.
func (c *Client) Reload(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodPut, "/configs?force=true", map[string]string{"path": path}, nil)
}
