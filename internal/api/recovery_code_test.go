package api

import (
	"strings"
	"testing"
)

func TestNormalizeRecoveryCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "normalizes with spaces and dashes",
			raw:  "  lume-abcd-2345-efgh  ",
			want: "LUME-ABCD-2345-EFGH",
		},
		{
			name: "normalizes raw 12 chars",
			raw:  "abcd2345efgh",
			want: "LUME-ABCD-2345-EFGH",
		},
		{
			name: "invalid length falls back to upper trimmed input",
			raw:  "abcd",
			want: "ABCD",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := normalizeRecoveryCode(test.raw)
			if got != test.want {
				t.Fatalf("normalizeRecoveryCode(%q) = %q, want %q", test.raw, got, test.want)
			}
		})
	}
}

func TestGenerateRecoveryCodeFormat(t *testing.T) {
	t.Parallel()

	code, err := generateRecoveryCode()
	if err != nil {
		t.Fatalf("generateRecoveryCode returned error: %v", err)
	}

	if !recoveryCodeRegex.MatchString(code) {
		t.Fatalf("generated code %q does not match required format", code)
	}

	if strings.ContainsAny(code, "IO10") {
		t.Fatalf("generated code %q contains ambiguous characters", code)
	}
}

func TestGenerateRecoveryCodeHash(t *testing.T) {
	t.Parallel()

	code, hash, err := generateRecoveryCodeHash()
	if err != nil {
		t.Fatalf("generateRecoveryCodeHash returned error: %v", err)
	}

	if !recoveryCodeRegex.MatchString(code) {
		t.Fatalf("generated code %q does not match required format", code)
	}
	if strings.TrimSpace(hash) == "" {
		t.Fatal("expected non-empty hash")
	}
}
