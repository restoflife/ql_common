package logger

import (
	"strings"
	"testing"
)

func TestTextLogPrefixesUseFixedWidth(t *testing.T) {
	prefixes := map[string]string{
		"xorm": XORM,
		"gorm": GORM,
		"gin":  GIN,
	}
	for name, prefix := range prefixes {
		if len(prefix) != 8 {
			t.Errorf("%s prefix width = %d, want 8", name, len(prefix))
		}
	}
}

func TestDatabaseLogPrefixesHaveNoPadding(t *testing.T) {
	prefixes := map[string]string{
		"sql":   SQL,
		"redis": REDIS,
		"mongo": MONGO,
	}
	for name, prefix := range prefixes {
		if strings.TrimSpace(prefix) != prefix {
			t.Errorf("%s prefix contains surrounding whitespace: %q", name, prefix)
		}
	}
	if SQL != "[mysql]" {
		t.Errorf("mysql prefix = %q, want %q", SQL, "[mysql]")
	}
}
