package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"

	"github.com/ZeroMarker/clashon/sdk/api"
)

func newTestModel() *Model {
	m := newModel(nil, "test")
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m.Update(proxiesMsg{
		"GLOBAL": {Name: "GLOBAL", Type: "Selector", Now: "一分机场", All: []string{"一分机场", "自动选择", "HK 01"}},
		"一分机场":   {Name: "一分机场", Type: "Selector", Now: "自动选择", All: []string{"自动选择", "HK 01", "US 02"}},
		"自动选择":   {Name: "自动选择", Type: "URLTest", Now: "HK 01", All: []string{"HK 01", "US 02"}},
		"HK 01":  {Name: "HK 01", Type: "Trojan", History: []struct{ Delay int `json:"delay"` }{{Delay: 120}}},
		"US 02":  {Name: "US 02", Type: "Shadowsocks"},
	})
	m.Update(configsMsg(&api.Configs{Mode: "rule"}))
	return m
}

func key(s string) tea.KeyMsg {
	switch s {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func checkFrame(t *testing.T, v string, w, h int) {
	t.Helper()
	lines := strings.Split(v, "\n")
	if len(lines) != h {
		t.Fatalf("frame has %d lines, want %d", len(lines), h)
	}
	for i, l := range lines {
		if got := lipglossWidth(l); got > w {
			t.Fatalf("line %d is %d cells wide (> %d): %q", i, got, w, l)
		}
	}
}

func TestProxiesView(t *testing.T) {
	m := newTestModel()
	if len(m.groups) != 3 || m.groups[0] != "一分机场" || m.groups[2] != "GLOBAL" {
		t.Fatalf("groups = %v", m.groups)
	}
	v := m.View()
	checkFrame(t, v, 80, 24)
	for _, want := range []string{"一分机场", "mode rule", "HK 01", "120ms"} {
		if !strings.Contains(v, want) {
			t.Errorf("view missing %q", want)
		}
	}
	// Enter focuses members at the current selection.
	m.Update(key("enter"))
	if !m.inMembers || m.pi != 0 {
		t.Fatalf("inMembers=%v pi=%d", m.inMembers, m.pi)
	}
}

func TestURLTestGroupRefusesManualSelect(t *testing.T) {
	m := newTestModel()
	m.Update(key("down")) // 自动选择
	m.Update(key("enter"))
	_, cmd := m.Update(key("enter"))
	if msg, ok := cmd().(statusMsg); !ok || !strings.Contains(string(msg), "URLTest") {
		t.Fatalf("got %v", msg)
	}
}

func TestTabsAndOtherViews(t *testing.T) {
	m := newTestModel()
	m.Update(rulesMsg{{Type: "DOMAIN-SUFFIX", Payload: "google.com", Proxy: "一分机场"}, {Type: "MATCH", Proxy: "DIRECT"}})
	m.Update(connsMsg(&api.Connections{Connections: []api.Connection{{ID: "1", Chains: []string{"HK 01", "一分机场"}, Start: time.Now()}}}))
	for i := 0; i < 50; i++ {
		m.Update(logMsg{Type: "info", Payload: strings.Repeat("长", 100)})
	}
	for tab, want := range map[string]string{"2": "一分机场 → HK 01", "3": "google.com", "4": "info"} {
		m.Update(key(tab))
		v := m.View()
		checkFrame(t, v, 80, 24)
		if !strings.Contains(v, want) {
			t.Errorf("tab %s missing %q", tab, want)
		}
	}
	m.Update(key("3"))
	m.Update(key("/"))
	for _, r := range "match" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if got := m.visibleRules(); len(got) != 1 || got[0].Type != "MATCH" {
		t.Fatalf("filtered = %v", got)
	}
}

func lipglossWidth(s string) int { return ansi.StringWidth(s) }
