package sysproxy

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Linux supports GNOME (gsettings) and KDE Plasma (kwriteconfig5/6).

func run(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s %s: %v: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func has(name string) bool { _, err := exec.LookPath(name); return err == nil }

func kwrite() string {
	for _, k := range []string{"kwriteconfig6", "kwriteconfig5"} {
		if has(k) {
			return k
		}
	}
	return ""
}

func gsettings(args ...string) error {
	_, err := run("gsettings", args...)
	return err
}

func enable(host string, port int) error {
	done := false
	if has("gsettings") {
		p := strconv.Itoa(port)
		ignore := "['" + strings.Join(Bypass, "', '") + "']"
		for _, a := range [][]string{
			{"set", "org.gnome.system.proxy.http", "host", host}, {"set", "org.gnome.system.proxy.http", "port", p},
			{"set", "org.gnome.system.proxy.https", "host", host}, {"set", "org.gnome.system.proxy.https", "port", p},
			{"set", "org.gnome.system.proxy.socks", "host", host}, {"set", "org.gnome.system.proxy.socks", "port", p},
			{"set", "org.gnome.system.proxy", "ignore-hosts", ignore},
			{"set", "org.gnome.system.proxy", "mode", "manual"},
		} {
			if err := gsettings(a...); err != nil {
				return err
			}
		}
		done = true
	}
	if k := kwrite(); k != "" {
		addr := fmt.Sprintf("http://%s %d", host, port)
		for _, kv := range [][2]string{
			{"httpProxy", addr}, {"httpsProxy", addr}, {"socksProxy", fmt.Sprintf("socks://%s %d", host, port)},
			{"NoProxyFor", strings.Join(Bypass, ",")}, {"ProxyType", "1"},
		} {
			if _, err := run(k, "--file", "kioslaverc", "--group", "Proxy Settings", "--key", kv[0], kv[1]); err != nil {
				return err
			}
		}
		done = true
	}
	if !done {
		return ErrUnsupported
	}
	return nil
}

func disable() error {
	done := false
	if has("gsettings") {
		if err := gsettings("set", "org.gnome.system.proxy", "mode", "none"); err != nil {
			return err
		}
		done = true
	}
	if k := kwrite(); k != "" {
		if _, err := run(k, "--file", "kioslaverc", "--group", "Proxy Settings", "--key", "ProxyType", "0"); err != nil {
			return err
		}
		done = true
	}
	if !done {
		return ErrUnsupported
	}
	return nil
}

func enabled(host string, port int) (bool, error) {
	if !has("gsettings") {
		return false, ErrUnsupported
	}
	mode, err := run("gsettings", "get", "org.gnome.system.proxy", "mode")
	if err != nil || mode != "'manual'" {
		return false, err
	}
	h, _ := run("gsettings", "get", "org.gnome.system.proxy.http", "host")
	p, _ := run("gsettings", "get", "org.gnome.system.proxy.http", "port")
	return h == "'"+host+"'" && p == strconv.Itoa(port), nil
}
