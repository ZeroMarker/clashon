package sysproxy

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const internetSettings = `Software\Microsoft\Windows\CurrentVersion\Internet Settings`

var (
	wininet           = windows.NewLazySystemDLL("wininet.dll")
	internetSetOption = wininet.NewProc("InternetSetOptionW")
)

// notify tells WinINet-based apps to reload proxy settings.
func notify() {
	const settingsChanged, refresh = 39, 37
	internetSetOption.Call(0, settingsChanged, 0, 0)
	internetSetOption.Call(0, refresh, 0, 0)
}

func enable(host string, port int) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, internetSettings, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if err := k.SetStringValue("ProxyServer", fmt.Sprintf("%s:%d", host, port)); err != nil {
		return err
	}
	// WinINet uses wildcards, not CIDR.
	bypass := []string{"localhost", "127.*", "10.*", "172.16.*", "192.168.*", "*.local", "<local>"}
	if err := k.SetStringValue("ProxyOverride", strings.Join(bypass, ";")); err != nil {
		return err
	}
	if err := k.SetDWordValue("ProxyEnable", 1); err != nil {
		return err
	}
	notify()
	return nil
}

func disable() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, internetSettings, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if err := k.SetDWordValue("ProxyEnable", 0); err != nil {
		return err
	}
	notify()
	return nil
}

func enabled(host string, port int) (bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, internetSettings, registry.QUERY_VALUE)
	if err != nil {
		return false, err
	}
	defer k.Close()
	on, _, err := k.GetIntegerValue("ProxyEnable")
	if err != nil || on == 0 {
		return false, nil
	}
	srv, _, _ := k.GetStringValue("ProxyServer")
	return srv == fmt.Sprintf("%s:%d", host, port), nil
}
