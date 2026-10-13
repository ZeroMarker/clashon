package sysproxy

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func networksetup(args ...string) (string, error) {
	out, err := exec.Command("networksetup", args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("networksetup %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

// services lists enabled network services (disabled ones start with "*").
func services() ([]string, error) {
	out, err := networksetup("-listallnetworkservices")
	if err != nil {
		return nil, err
	}
	var s []string
	for _, l := range strings.Split(out, "\n")[1:] {
		if l = strings.TrimSpace(l); l != "" && !strings.HasPrefix(l, "*") {
			s = append(s, l)
		}
	}
	return s, nil
}

func enable(host string, port int) error {
	svcs, err := services()
	if err != nil {
		return err
	}
	p := strconv.Itoa(port)
	for _, s := range svcs {
		for _, kind := range []string{"webproxy", "securewebproxy", "socksfirewallproxy"} {
			if _, err := networksetup("-set"+kind, s, host, p); err != nil {
				return err
			}
		}
		if _, err := networksetup(append([]string{"-setproxybypassdomains", s}, Bypass...)...); err != nil {
			return err
		}
	}
	return nil
}

func disable() error {
	svcs, err := services()
	if err != nil {
		return err
	}
	for _, s := range svcs {
		for _, kind := range []string{"webproxystate", "securewebproxystate", "socksfirewallproxystate"} {
			if _, err := networksetup("-set"+kind, s, "off"); err != nil {
				return err
			}
		}
	}
	return nil
}

func enabled(host string, port int) (bool, error) {
	svcs, err := services()
	if err != nil || len(svcs) == 0 {
		return false, err
	}
	out, err := networksetup("-getwebproxy", svcs[0])
	if err != nil {
		return false, err
	}
	return strings.Contains(out, "Enabled: Yes") && strings.Contains(out, "Server: "+host) &&
		strings.Contains(out, "Port: "+strconv.Itoa(port)), nil
}
