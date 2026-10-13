package desktop

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLinuxIntegration(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(dir, "data"))
	t.Setenv("PATH", "") // skip xdg-mime/update-desktop-database
	exe := "/opt/Clash On/clashon-gui"

	if on, _ := Autostart(); on {
		t.Fatal("autostart should start off")
	}
	if err := SetAutostart(exe, true); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(autostartFile())
	if !strings.Contains(string(b), `Exec="/opt/Clash On/clashon-gui" --minimized`) {
		t.Fatalf("autostart entry:\n%s", b)
	}
	if on, _ := Autostart(); !on {
		t.Fatal("autostart should be on")
	}
	if err := SetAutostart(exe, false); err != nil {
		t.Fatal(err)
	}
	if on, _ := Autostart(); on {
		t.Fatal("autostart should be off")
	}

	if err := RegisterSchemes(exe); err != nil {
		t.Fatal(err)
	}
	b, _ = os.ReadFile(filepath.Join(dir, "data", "applications", appID+".desktop"))
	if !strings.Contains(string(b), "MimeType=x-scheme-handler/clash;x-scheme-handler/clashmeta;x-scheme-handler/mihomo;") ||
		!strings.Contains(string(b), " %u") {
		t.Fatalf("scheme entry:\n%s", b)
	}
}
