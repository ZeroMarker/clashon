package deeplink

import "testing"

func TestParse(t *testing.T) {
	for _, tc := range []struct{ in, url, name string }{
		{"clash://install-config?url=https%3A%2F%2Fa.com%2Fsub%3Ftoken%3Dx&name=My%20Air", "https://a.com/sub?token=x", "My-Air"},
		{"clashmeta://install-config?url=https://b.org/s", "https://b.org/s", "b.org"},
		{"clash://install-config?url=https://c.net/s&name=../../etc", "https://c.net/s", "etc"},
	} {
		got, err := Parse(tc.in)
		if err != nil {
			t.Fatalf("%s: %v", tc.in, err)
		}
		if got.URL != tc.url || got.Name != tc.name {
			t.Errorf("%s: got %+v", tc.in, got)
		}
	}
	for _, bad := range []string{"http://install-config?url=https://a", "clash://other?url=https://a", "clash://install-config?url=file:///etc/passwd"} {
		if _, err := Parse(bad); err == nil {
			t.Errorf("%s: expected error", bad)
		}
	}
}
