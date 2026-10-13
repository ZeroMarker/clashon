// Package config merges a subscription profile with local overrides into the
// runtime config fed to mihomo.
package config

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Settings are user-level overrides applied on top of every profile.
type Settings struct {
	MixedPort  int    `json:"mixed-port"`
	Controller string `json:"external-controller"`
	Secret     string `json:"secret"`
	AllowLAN   bool   `json:"allow-lan"`
	Mode       string `json:"mode,omitempty"`      // rule | global | direct; empty keeps profile value
	LogLevel   string `json:"log-level,omitempty"` // empty keeps profile value
	TUN        bool   `json:"tun"`                 // route all traffic via a TUN device; needs privileges
}

// Defaults returns settings with a fresh random controller secret.
func Defaults() Settings {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return Settings{MixedPort: 7890, Controller: "127.0.0.1:9090", Secret: hex.EncodeToString(b)}
}

// LoadSettings reads path, creating it with Defaults if missing.
func LoadSettings(path string) (Settings, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		s := Defaults()
		return s, SaveSettings(path, s)
	}
	if err != nil {
		return Settings{}, err
	}
	s := Defaults()
	if err := json.Unmarshal(b, &s); err != nil {
		return Settings{}, fmt.Errorf("%s: %w", path, err)
	}
	return s, nil
}

// SaveSettings writes s to path.
func SaveSettings(path string, s Settings) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

// Overridden top-level keys that conflict with ClashOn's own listeners.
var dropped = []string{"port", "socks-port", "redir-port", "tproxy-port", "external-controller-tls", "external-controller-unix", "external-ui"}

// Merge applies s to the profile YAML and returns the runtime config.
func Merge(profile []byte, s Settings) ([]byte, error) {
	var doc map[string]any
	if err := yaml.Unmarshal(profile, &doc); err != nil {
		return nil, fmt.Errorf("parse profile: %w", err)
	}
	if doc == nil {
		doc = map[string]any{}
	}
	for _, k := range dropped {
		delete(doc, k)
	}
	doc["mixed-port"] = s.MixedPort
	doc["external-controller"] = s.Controller
	doc["secret"] = s.Secret
	doc["allow-lan"] = s.AllowLAN
	if s.Mode != "" {
		doc["mode"] = s.Mode
	}
	if s.LogLevel != "" {
		doc["log-level"] = s.LogLevel
	}
	if s.TUN {
		applyTUN(doc)
	} else {
		delete(doc, "tun")
	}
	return yaml.Marshal(doc)
}

// applyTUN enables a TUN device and the DNS hijacking it depends on.
// Profile values for keys we do not need to force are kept.
func applyTUN(doc map[string]any) {
	tun, _ := doc["tun"].(map[string]any)
	if tun == nil {
		tun = map[string]any{"stack": "mixed"}
	}
	tun["enable"] = true
	tun["auto-route"] = true
	tun["auto-detect-interface"] = true
	if _, ok := tun["dns-hijack"]; !ok {
		tun["dns-hijack"] = []string{"any:53", "tcp://any:53"}
	}
	doc["tun"] = tun

	dns, _ := doc["dns"].(map[string]any)
	if dns == nil {
		dns = map[string]any{
			"enhanced-mode": "fake-ip",
			"nameserver":    []string{"https://223.5.5.5/dns-query", "https://1.1.1.1/dns-query"},
		}
	}
	dns["enable"] = true
	doc["dns"] = dns
}
