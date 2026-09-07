package registry

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestAtomicBatchAndCleanup(t *testing.T) {
	var r Registry[string]
	failure := errors.New("build failure")
	closed := make(map[string]int)
	closeItem := func(s string) error { closed[s]++; return nil }
	if err := r.Init([]string{"existing"}, func(s string) (string, error) { return s, nil }, closeItem); err != nil {
		t.Fatal(err)
	}
	err := r.Init([]string{"a", "b"}, func(s string) (string, error) {
		if _, err := r.Get("a"); !errors.Is(err, ErrNotFound) {
			t.Fatal("partial batch published")
		}
		if s == "b" {
			return "", failure
		}
		return s, nil
	}, closeItem)
	if !errors.Is(err, failure) || closed["a"] != 1 {
		t.Fatalf("error=%v closed=%v", err, closed)
	}
	if _, err := r.Get("a"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	if _, err := r.Get("existing"); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(closeItem); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(closeItem); err != nil {
		t.Fatal(err)
	}
	if closed["existing"] != 1 {
		t.Fatal(closed)
	}
	if err := r.Init([]string{"existing"}, func(s string) (string, error) { return s, nil }, closeItem); err != nil {
		t.Fatal(err)
	}
}

func TestDuplicateValidationAndCleanupErrors(t *testing.T) {
	var r Registry[int]
	calls := 0
	build := func(string) (int, error) { calls++; return 1, nil }
	closeItem := func(int) error { return nil }
	for _, names := range [][]string{{""}, {"a", "a"}} {
		if err := r.Init(names, build, closeItem); err == nil {
			t.Fatal("expected validation error")
		}
	}
	if calls != 0 {
		t.Fatal("built invalid batch")
	}
	if err := r.Init([]string{"a"}, build, closeItem); err != nil {
		t.Fatal(err)
	}
	if err := r.Init([]string{"a", "b"}, build, closeItem); !errors.Is(err, ErrDuplicate) {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal(calls)
	}
	closeErr := errors.New("close error")
	if err := r.Close(func(int) error { return closeErr }); !errors.Is(err, closeErr) {
		t.Fatal(err)
	}
	if _, err := r.Get("a"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
}

func TestConcurrentLifecycleAndLookup(t *testing.T) {
	var r Registry[int]
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				name := fmt.Sprintf("%d-%d", i, j)
				if err := r.Init([]string{name}, func(string) (int, error) { return i, nil }, func(int) error { return nil }); err != nil {
					t.Error(err)
				}
				_, _ = r.Get(name)
				if err := r.Close(func(int) error { return nil }); err != nil {
					t.Error(err)
				}
			}
		}(i)
	}
	wg.Wait()
}

func TestPanicCleansPreviouslyBuiltResources(t *testing.T) {
	var r Registry[int]
	closed := 0
	func() {
		defer func() {
			if recover() != "boom" {
				t.Error("panic not propagated")
			}
		}()
		_ = r.Init([]string{"a", "b"}, func(s string) (int, error) {
			if s == "b" {
				panic("boom")
			}
			return 1, nil
		}, func(int) error { closed++; return nil })
	}()
	if closed != 1 {
		t.Fatal(closed)
	}
	if _, err := r.Get("a"); !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
}
