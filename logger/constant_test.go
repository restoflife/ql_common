package logger

import (
	"testing"
)

func TestLogPrefixesHaveNoPadding(t *testing.T) {
	prefixes := map[string]string{
		"xorm":  "[xorm]",
		"gorm":  "[gorm]",
		"gin":   "[gin]",
		"sql":   "[mysql]",
		"redis": "[redis]",
		"mongo": "[mongo]",
	}
	actual := map[string]string{
		"xorm":  XORM,
		"gorm":  GORM,
		"gin":   GIN,
		"sql":   SQL,
		"redis": REDIS,
		"mongo": MONGO,
	}
	for name, expected := range prefixes {
		if actual[name] != expected {
			t.Errorf("%s prefix = %q, want %q", name, actual[name], expected)
		}
	}
}
