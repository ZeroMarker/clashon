package desktop

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func dataHome() string {
	if d := os.Getenv("XDG_DATA_HOME"); d != "" {
		return d
	}
	h, _ := os.UserHomeDir()
	return filepath.Join(h, ".local", "share")
}

func configHome() string {
	d, _ := os.UserConfigDir()
	return d
}

func autostartFile() string { return filepath.Join(configHome(), "autostart", appID+".desktop") }

func desktopEntry(exe string, extra ...string) string {
	lines := []string{
		"[Desktop Entry]",
		"Type=Application",
		"Name=" + appName,
		"Exec=" + quote(exe) + " %u",
		"Icon=" + appID,
		"Categories=Network;",
		"Terminal=false",
	}
	return strings.Join(append(lines, extra...), "\n") + "\n"
}

// quote escapes exe for the Exec key per the Desktop Entry spec.
func quote(s string) string {
	if !strings.ContainsAny(s, " \t\"'\\$`") {
		return s
	}
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "`", "\\`", "$", `\$`)
	return `"` + r.Replace(s) + `"`
}

func write(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func setAutostart(exe string, on bool) error {
	if !on {
		if err := os.Remove(autostartFile()); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	return write(autostartFile(), strings.Replace(desktopEntry(exe), " %u", " --minimized", 1))
}

func autostart() (bool, error) {
	_, err := os.Stat(autostartFile())
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func registerSchemes(exe string) error {
	var mime []string
	for _, s := range Schemes {
		mime = append(mime, "x-scheme-handler/"+s)
	}
	dir := filepath.Join(dataHome(), "applications")
	file := appID + ".desktop"
	if err := write(filepath.Join(dir, file), desktopEntry(exe, "MimeType="+strings.Join(mime, ";")+";")); err != nil {
		return err
	}
	if p, err := exec.LookPath("update-desktop-database"); err == nil {
		_ = exec.Command(p, dir).Run()
	}
	if p, err := exec.LookPath("xdg-mime"); err == nil {
		for _, m := range mime {
			if out, err := exec.Command(p, "default", file, m).CombinedOutput(); err != nil {
				return fmt.Errorf("xdg-mime: %v: %s", err, out)
			}
		}
	}
	return nil
}
