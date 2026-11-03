package profile

import "testing"

func TestValidateName(t *testing.T) {
	for _, ok := range []string{"default", "一分机场", "My Air", "机场 2", "a.b-c_d", "日本🇯🇵"} {
		if err := ValidateName(ok); err != nil {
			t.Errorf("%q: %v", ok, err)
		}
	}
	for _, bad := range []string{"", "..", "a/b", `a\b`, "a:b", "con", "COM1.yaml", " lead", "trail.", "a\x00b"} {
		if ValidateName(bad) == nil {
			t.Errorf("%q should be rejected", bad)
		}
	}
}

func TestSanitizeName(t *testing.T) {
	for in, want := range map[string]string{
		"一分机场":       "一分机场",
		"../../etc":  "etc",
		"a/b:c":      "a-b-c",
		"  空格  ":     "空格",
		"CON":        "",
		"///":        "",
	} {
		if got := SanitizeName(in); got != want {
			t.Errorf("SanitizeName(%q) = %q, want %q", in, got, want)
		}
	}
}
