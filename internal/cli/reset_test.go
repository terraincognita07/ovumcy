package cli

import (
	"strings"
	"testing"
)

func TestGenerateTemporaryPasswordMinimumLength(t *testing.T) {
	t.Parallel()

	password, err := generateTemporaryPassword(4)
	if err != nil {
		t.Fatalf("generateTemporaryPassword returned error: %v", err)
	}
	if len(password) != 8 {
		t.Fatalf("generateTemporaryPassword minimum len = %d, want 8", len(password))
	}
}

func TestGenerateTemporaryPasswordAlphabet(t *testing.T) {
	t.Parallel()

	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"
	password, err := generateTemporaryPassword(24)
	if err != nil {
		t.Fatalf("generateTemporaryPassword returned error: %v", err)
	}
	if len(password) != 24 {
		t.Fatalf("generateTemporaryPassword len = %d, want 24", len(password))
	}

	for _, char := range password {
		if !strings.ContainsRune(alphabet, char) {
			t.Fatalf("password %q contains char %q outside alphabet", password, char)
		}
	}
}
