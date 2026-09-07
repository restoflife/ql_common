package db

import (
	"errors"
	"fmt"
	"github.com/restoflife/ql_common/internal/registry"
)

var (
	ErrNotFound               = registry.ErrNotFound
	ErrDuplicate              = registry.ErrDuplicate
	ErrNilContext             = errors.New("context must not be nil")
	ErrNilTransactionCallback = errors.New("transaction callback must not be nil")
	ErrInvalidDatabaseConfig  = errors.New("invalid database configuration")
)

func databaseConfigError(name, detail string) error {
	return fmt.Errorf("%w for %q: %s", ErrInvalidDatabaseConfig, name, detail)
}
