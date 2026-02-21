package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func TestLocaleKeysParity(t *testing.T) {
	en := mustLoadLocaleMessages(t, "en")
	ru := mustLoadLocaleMessages(t, "ru")

	missingInRU := missingKeys(en, ru)
	missingInEN := missingKeys(ru, en)

	if len(missingInRU) == 0 && len(missingInEN) == 0 {
		return
	}

	if len(missingInRU) > 0 {
		t.Errorf("keys missing in ru locale: %s", strings.Join(missingInRU, ", "))
	}
	if len(missingInEN) > 0 {
		t.Errorf("keys missing in en locale: %s", strings.Join(missingInEN, ", "))
	}
}

func mustLoadLocaleMessages(t *testing.T, language string) map[string]string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("resolve test file path: runtime.Caller failed")
	}
	localesDir := filepath.Join(filepath.Dir(thisFile), "locales")
	localePath := filepath.Join(localesDir, language+".json")

	content, err := os.ReadFile(localePath)
	if err != nil {
		t.Fatalf("read locale %q: %v", language, err)
	}

	messages := map[string]string{}
	if err := json.Unmarshal(content, &messages); err != nil {
		t.Fatalf("parse locale %q: %v", language, err)
	}
	if len(messages) == 0 {
		t.Fatalf("locale %q is empty", language)
	}

	return messages
}

func missingKeys(source map[string]string, target map[string]string) []string {
	missing := make([]string, 0)
	for key := range source {
		if _, ok := target[key]; !ok {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	return missing
}
