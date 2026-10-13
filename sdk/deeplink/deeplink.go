// Package deeplink parses clash://install-config links.
package deeplink

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Install is a request to add a subscription.
type Install struct {
	URL  string
	Name string
}

var unsafe = regexp.MustCompile(`[^\w.-]+`)

// Parse accepts clash://install-config?url=<escaped>&name=<name>
// (also clashmeta:// and mihomo://). Name defaults to the subscription host.
func Parse(link string) (*Install, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "clash", "clashmeta", "mihomo":
	default:
		return nil, fmt.Errorf("unsupported scheme %q", u.Scheme)
	}
	if u.Host != "install-config" && strings.TrimPrefix(u.Opaque, "//") != "install-config" {
		return nil, fmt.Errorf("unsupported action %q", u.Host)
	}
	q := u.Query()
	sub := q.Get("url")
	su, err := url.Parse(sub)
	if err != nil || (su.Scheme != "http" && su.Scheme != "https") || su.Host == "" {
		return nil, fmt.Errorf("invalid subscription url %q", sub)
	}
	name := q.Get("name")
	if name == "" {
		name = su.Hostname()
	}
	name = strings.Trim(unsafe.ReplaceAllString(name, "-"), "-.")
	if name == "" {
		name = "imported"
	}
	return &Install{URL: sub, Name: name}, nil
}
