package redis

import (
	"errors"
	"fmt"
	driver "github.com/redis/go-redis/v9"
	"github.com/restoflife/ql_common/internal/registry"
)

var (
	Nil                = driver.Nil
	ErrNotFound        = registry.ErrNotFound
	ErrDuplicate       = registry.ErrDuplicate
	ErrNilContext      = errors.New("context must not be nil")
	ErrNilPipeline     = errors.New("pipeline callback must not be nil")
	ErrCommandRequired = errors.New("Redis command is required")
	ErrInvalidConfig   = errors.New("invalid Redis configuration")
)

func configError(name, detail string) error {
	return fmt.Errorf("%w for %q: %s", ErrInvalidConfig, name, detail)
}
