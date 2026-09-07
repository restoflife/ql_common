package mongo

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

// Exercise every Context entry point, including variadic APIs, without a server.
func TestAllContextEntryPoints(t *testing.T) {
	cases := []struct {
		name string
		fn   any
	}{
		{"RunCommandContext", RunCommandContext},
		{"InsertOneResultContext", InsertOneResultContext},
		{"InsertOneContext", InsertOneContext},
		{"InsertManyContext", InsertManyContext},
		{"FindOneContext", FindOneContext},
		{"FindContext", FindContext},
		{"UpdateOneContext", UpdateOneContext},
		{"UpdateManyContext", UpdateManyContext},
		{"UpdateByIDContext", UpdateByIDContext},
		{"DeleteOneContext", DeleteOneContext},
		{"DeleteManyContext", DeleteManyContext},
		{"CountDocumentsContext", CountDocumentsContext},
		{"EstimatedDocumentCountContext", EstimatedDocumentCountContext},
		{"AggregateContext", AggregateContext},
		{"DistinctContext", DistinctContext},
		{"DropIndexContext", DropIndexContext},
		{"CreateIndexContext", CreateIndexContext},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fn := reflect.ValueOf(tc.fn)
			typ := fn.Type()
			args := make([]reflect.Value, typ.NumIn())
			for i := range args {
				args[i] = reflect.Zero(typ.In(i))
				if typ.In(i).Kind() == reflect.Func {
					ft := typ.In(i)
					args[i] = reflect.MakeFunc(ft, func([]reflect.Value) []reflect.Value {
						out := make([]reflect.Value, ft.NumOut())
						for j := range out {
							out[j] = reflect.Zero(ft.Out(j))
						}
						return out
					})
				}
			}
			call := func() error {
				var result []reflect.Value
				if typ.IsVariadic() {
					result = fn.CallSlice(args)
				} else {
					result = fn.Call(args)
				}
				last := result[len(result)-1]
				if last.IsNil() {
					return nil
				}
				return last.Interface().(error)
			}
			if err := call(); err == nil {
				t.Fatal("nil context accepted")
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			args[0] = reflect.ValueOf(ctx)
			if err := call(); !errors.Is(err, context.Canceled) {
				t.Fatalf("cancellation lost: %v", err)
			}
			args[0] = reflect.ValueOf(context.Background())
			if err := call(); !errors.Is(err, ErrNotFound) {
				t.Fatalf("missing instance error lost: %v", err)
			}
		})
	}
}
