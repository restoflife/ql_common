package mongo

import (
	"errors"
	"fmt"
	"github.com/restoflife/ql_common/internal/registry"
)

var (
	ErrNotFound             = registry.ErrNotFound
	ErrDuplicate            = registry.ErrDuplicate
	ErrNilContext           = errors.New("context must not be nil")
	ErrInvalidConfig        = errors.New("invalid MongoDB configuration")
	ErrURIRequired          = errors.New("MongoDB URI is required")
	ErrPoolSize             = errors.New("MongoDB min_pool_size exceeds max_pool_size")
	ErrUsernameRequired     = errors.New("MongoDB username is required with password")
	ErrNegativeTimeout      = errors.New("MongoDB timeout must not be negative")
	ErrDatabaseRequired     = errors.New("MongoDB database is required")
	ErrCollectionRequired   = errors.New("MongoDB collection name is invalid")
	ErrInvalidCACertificate = errors.New("cannot parse CA certificate")
	ErrObjectIDRequired     = errors.New("InsertOne requires an ObjectID _id; use GetCollection for custom IDs")
	ErrUnexpectedInsertID   = errors.New("insert succeeded but returned a non-ObjectID; do not retry blindly")
)

func configError(name string, err error) error {
	return fmt.Errorf("%w for %q: %w", ErrInvalidConfig, name, err)
}

func operationError(operation string, err error) error {
	return fmt.Errorf("MongoDB %s failed: %w", operation, err)
}
