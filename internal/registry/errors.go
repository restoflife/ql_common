package registry

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrDuplicate    = errors.New("resource already exists")
	ErrNameRequired = errors.New("resource name must not be empty")
)

func duplicateError(name string) error { return fmt.Errorf("%w: %s", ErrDuplicate, name) }
func notFoundError(name string) error  { return fmt.Errorf("%w: %s", ErrNotFound, name) }
func lifecycleError(action, name string, err error) error {
	return fmt.Errorf("%s %s: %w", action, name, err)
}
