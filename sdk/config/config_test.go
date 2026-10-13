package config

import (
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMerge(t *testing.T) {
	in := []byte("port: 7891\nsocks-port: 7892\nexternal-controller: 0.0.0.0:9999\nmode: rule\nproxies: []\n")
	s := Settings{MixedPort: 1080, Controller: "127.0.0.1:9090", Secret: "s", Mode: "global"}
	out, err := Merge(in, s)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := yaml.Unmarshal(out, &doc); err != nil {
		t.Fatal(err)
	}
	if doc["mixed-port"] != 1080 || doc["external-controller"] != "127.0.0.1:9090" || doc["secret"] != "s" || doc["mode"] != "global" {
		t.Fatalf("overrides not applied: %v", doc)
	}
	if _, ok := doc["port"]; ok {
		t.Fatal("port should be dropped")
	}
	if _, ok := doc["proxies"]; !ok {
		t.Fatal("proxies lost")
	}
}

func TestLoadSettingsCreatesDefaults(t *testing.T) {
	p := filepath.Join(t.TempDir(), "settings.json")
	a, err := LoadSettings(p)
	if err != nil {
		t.Fatal(err)
	}
	b, err := LoadSettings(p)
	if err != nil {
		t.Fatal(err)
	}
	if a.Secret == "" || a.Secret != b.Secret {
		t.Fatalf("secret not persisted: %q vs %q", a.Secret, b.Secret)
	}
}
