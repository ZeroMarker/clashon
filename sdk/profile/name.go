package profile

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Profile names become file names, so they must be valid on every OS.
// Any language is fine; path separators, characters Windows forbids,
// control characters and reserved device names are not.

const maxNameLen = 64

const forbidden = `<>:"/\|?*`

var reserved = map[string]bool{"CON": true, "PRN": true, "AUX": true, "NUL": true}

func init() {
	for i := 1; i <= 9; i++ {
		reserved[fmt.Sprintf("COM%d", i)] = true
		reserved[fmt.Sprintf("LPT%d", i)] = true
	}
}

// ValidateName reports why name cannot be used as a profile name.
func ValidateName(name string) error {
	switch {
	case name == "":
		return fmt.Errorf("profile name is empty")
	case utf8.RuneCountInString(name) > maxNameLen:
		return fmt.Errorf("profile name is longer than %d characters", maxNameLen)
	case !utf8.ValidString(name):
		return fmt.Errorf("profile name is not valid UTF-8")
	case strings.TrimSpace(name) != name || strings.HasSuffix(name, "."):
		return fmt.Errorf("profile name %q must not start or end with a space or end with a dot", name)
	case name == "." || name == "..":
		return fmt.Errorf("invalid profile name %q", name)
	}
	for _, r := range name {
		if strings.ContainsRune(forbidden, r) || unicode.IsControl(r) {
			return fmt.Errorf("profile name %q must not contain %q", name, r)
		}
	}
	stem, _, _ := strings.Cut(name, ".")
	if reserved[strings.ToUpper(stem)] {
		return fmt.Errorf("profile name %q is reserved on Windows", name)
	}
	return nil
}

// SanitizeName turns arbitrary text (e.g. from a clash:// link) into a valid
// name, or returns "" if nothing usable is left.
func SanitizeName(s string) string {
	var b strings.Builder
	for _, r := range s {
		if strings.ContainsRune(forbidden, r) || unicode.IsControl(r) || r == utf8.RuneError {
			r = '-'
		}
		b.WriteRune(r)
	}
	out := strings.Trim(b.String(), " .-")
	if r := []rune(out); len(r) > maxNameLen {
		out = strings.TrimRight(string(r[:maxNameLen]), " .")
	}
	if ValidateName(out) != nil {
		return ""
	}
	return out
}
