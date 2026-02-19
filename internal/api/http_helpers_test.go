package api

import "testing"

func TestSanitizeRedirectPath(t *testing.T) {
	t.Parallel()

	fallback := "/login"

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "empty uses fallback", raw: "", want: fallback},
		{name: "absolute url blocked", raw: "https://evil.example", want: fallback},
		{name: "protocol relative blocked", raw: "//evil.example", want: fallback},
		{name: "path without leading slash blocked", raw: "dashboard", want: fallback},
		{name: "valid local path kept", raw: "/dashboard", want: "/dashboard"},
		{name: "valid local path with query kept", raw: "/calendar?month=2026-02&day=2026-02-17", want: "/calendar?month=2026-02&day=2026-02-17"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := sanitizeRedirectPath(test.raw, fallback)
			if got != test.want {
				t.Fatalf("sanitizeRedirectPath(%q) = %q, want %q", test.raw, got, test.want)
			}
		})
	}
}
