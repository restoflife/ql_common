package redis

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
		{"ZRevRangeContext", ZRevRangeContext},
		{"ZRevRangeWithScoresContext", ZRevRangeWithScoresContext},
		{"SetContext", SetContext},
		{"SetNXContext", SetNXContext},
		{"SetXXContext", SetXXContext},
		{"GetContext", GetContext},
		{"GetDelContext", GetDelContext},
		{"GetExContext", GetExContext},
		{"IncrContext", IncrContext},
		{"IncrByContext", IncrByContext},
		{"IncrByFloatContext", IncrByFloatContext},
		{"DecrContext", DecrContext},
		{"DecrByContext", DecrByContext},
		{"AppendContext", AppendContext},
		{"MGetContext", MGetContext},
		{"MSetContext", MSetContext},
		{"GetRangeContext", GetRangeContext},
		{"SetRangeContext", SetRangeContext},
		{"StrLenContext", StrLenContext},
		{"HSetContext", HSetContext},
		{"HSetNXContext", HSetNXContext},
		{"HGetContext", HGetContext},
		{"HGetAllContext", HGetAllContext},
		{"HMSetContext", HMSetContext},
		{"HMGetContext", HMGetContext},
		{"HDelContext", HDelContext},
		{"HExistsContext", HExistsContext},
		{"HIncrByContext", HIncrByContext},
		{"HIncrByFloatContext", HIncrByFloatContext},
		{"HKeysContext", HKeysContext},
		{"HLenContext", HLenContext},
		{"HValsContext", HValsContext},
		{"LPushContext", LPushContext},
		{"RPushContext", RPushContext},
		{"LPopContext", LPopContext},
		{"RPopContext", RPopContext},
		{"LRangeContext", LRangeContext},
		{"LLenContext", LLenContext},
		{"LIndexContext", LIndexContext},
		{"LInsertContext", LInsertContext},
		{"LRemContext", LRemContext},
		{"LTrimContext", LTrimContext},
		{"LSetContext", LSetContext},
		{"SAddContext", SAddContext},
		{"SRemContext", SRemContext},
		{"SMembersContext", SMembersContext},
		{"SIsMemberContext", SIsMemberContext},
		{"SCardContext", SCardContext},
		{"SPopContext", SPopContext},
		{"SUnionContext", SUnionContext},
		{"SInterContext", SInterContext},
		{"SDiffContext", SDiffContext},
		{"ZAddContext", ZAddContext},
		{"ZRemContext", ZRemContext},
		{"ZRangeContext", ZRangeContext},
		{"ZRangeWithScoresContext", ZRangeWithScoresContext},
		{"ZRankContext", ZRankContext},
		{"ZScoreContext", ZScoreContext},
		{"ZIncrByContext", ZIncrByContext},
		{"ZCardContext", ZCardContext},
		{"ZCountContext", ZCountContext},
		{"ZRemRangeByRankContext", ZRemRangeByRankContext},
		{"ZRemRangeByScoreContext", ZRemRangeByScoreContext},
		{"ExistsContext", ExistsContext},
		{"DelContext", DelContext},
		{"ExpireContext", ExpireContext},
		{"ExpireAtContext", ExpireAtContext},
		{"TTLContext", TTLContext},
		{"PersistContext", PersistContext},
		{"RenameContext", RenameContext},
		{"RenameNXContext", RenameNXContext},
		{"TypeContext", TypeContext},
		{"KeysContext", KeysContext},
		{"ScanContext", ScanContext},
		{"PipelineContext", PipelineContext},
		{"TxPipelineContext", TxPipelineContext},
		{"EvalContext", EvalContext},
		{"EvalShaContext", EvalShaContext},
		{"ScriptLoadContext", ScriptLoadContext},
		{"ScriptExistsContext", ScriptExistsContext},
		{"PFAddContext", PFAddContext},
		{"PFCountContext", PFCountContext},
		{"PFMergeContext", PFMergeContext},
		{"SetBitContext", SetBitContext},
		{"GetBitContext", GetBitContext},
		{"BitCountContext", BitCountContext},
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
