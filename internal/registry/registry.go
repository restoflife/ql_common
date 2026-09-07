// Package registry manages named resources with atomic batch publication.
package registry

import (
	"errors"
	"sort"
	"sync"
)

// Registry's zero value is ready to use. Lifecycle operations are serialized;
// lookups never hold a lock during network I/O. Do not copy after first use.
type Registry[T any] struct {
	lifecycle sync.Mutex
	mu        sync.RWMutex
	items     map[string]T
}

// Init publishes all resources together, or closes everything built by this call.
// build must clean up its own resource when returning an error. Callbacks must
// not recursively call Init or Close on the same registry.
func (r *Registry[T]) Init(names []string, build func(string) (T, error), closeResource func(T) error) (err error) {
	r.lifecycle.Lock()
	defer r.lifecycle.Unlock()
	names = append([]string(nil), names...)
	sort.Strings(names)
	r.mu.RLock()
	for i, name := range names {
		if name == "" {
			r.mu.RUnlock()
			return ErrNameRequired
		}
		_, exists := r.items[name]
		if exists || (i > 0 && names[i-1] == name) {
			r.mu.RUnlock()
			return duplicateError(name)
		}
	}
	r.mu.RUnlock()
	pending := make(map[string]T, len(names))
	committed := false
	defer func() {
		if !committed {
			for name, item := range pending {
				if closeErr := closeResource(item); closeErr != nil {
					err = errors.Join(err, lifecycleError("cleanup", name, closeErr))
				}
			}
		}
	}()
	for _, name := range names {
		item, buildErr := build(name)
		if buildErr != nil {
			return lifecycleError("initialize", name, buildErr)
		}
		pending[name] = item
	}
	r.mu.Lock()
	if r.items == nil {
		r.items = make(map[string]T)
	}
	for name, item := range pending {
		r.items[name] = item
	}
	r.mu.Unlock()
	committed = true
	return nil
}

func (r *Registry[T]) Get(name string) (T, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[name]
	if !ok {
		return item, notFoundError(name)
	}
	return item, nil
}

// Close removes entries before closing them and is safe to call repeatedly.
// The application must drain in-flight users before calling Close.
func (r *Registry[T]) Close(closeResource func(T) error) error {
	r.lifecycle.Lock()
	defer r.lifecycle.Unlock()
	r.mu.Lock()
	items := r.items
	r.items = nil
	r.mu.Unlock()
	var result error
	for name, item := range items {
		if err := closeResource(item); err != nil {
			result = errors.Join(result, lifecycleError("close", name, err))
		}
	}
	return result
}
