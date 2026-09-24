package logger

import "testing"

func TestLogPrefixesUseFixedWidth(t *testing.T) {
	prefixes := map[string]string{
		"xorm":  XORM,
		"gorm":  GORM,
		"sql":   SQL,
		"redis": REDIS,
		"mongo": MONGO,
		"gin":   GIN,
	}
	for name, prefix := range prefixes {
		if len(prefix) != 8 {
			t.Errorf("%s prefix width = %d, want 8", name, len(prefix))
		}
	}
}
