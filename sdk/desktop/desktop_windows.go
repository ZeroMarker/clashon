package desktop

import (
	"golang.org/x/sys/windows/registry"
)

const runKey = `Software\Microsoft\Windows\CurrentVersion\Run`

func setAutostart(exe string, on bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if !on {
		if err := k.DeleteValue(appName); err != nil && err != registry.ErrNotExist {
			return err
		}
		return nil
	}
	return k.SetStringValue(appName, `"`+exe+`" --minimized`)
}

func autostart() (bool, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKey, registry.QUERY_VALUE)
	if err != nil {
		return false, err
	}
	defer k.Close()
	_, _, err = k.GetStringValue(appName)
	if err == registry.ErrNotExist {
		return false, nil
	}
	return err == nil, err
}

func registerSchemes(exe string) error {
	for _, s := range Schemes {
		k, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Classes\`+s, registry.SET_VALUE)
		if err != nil {
			return err
		}
		k.SetStringValue("", "URL:"+s+" protocol")
		k.SetStringValue("URL Protocol", "")
		k.Close()
		cmd, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Classes\`+s+`\shell\open\command`, registry.SET_VALUE)
		if err != nil {
			return err
		}
		err = cmd.SetStringValue("", `"`+exe+`" "%1"`)
		cmd.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
