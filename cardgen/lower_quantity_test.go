package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
)

func TestPositiveQuantityParameterBounds(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		amount  compiler.CompiledAmount
		dynamic bool
		ok      bool
	}{
		{"fixed", compiler.CompiledAmount{Known: true, Value: 2}, false, true},
		{"zero", compiler.CompiledAmount{Known: true}, true, false},
		{"negative", compiler.CompiledAmount{Known: true, Value: -1}, true, false},
		{"X", compiler.CompiledAmount{VariableX: true}, true, true},
		{"X not allowed", compiler.CompiledAmount{VariableX: true}, false, false},
		{"unknown", compiler.CompiledAmount{}, true, false},
		{"source power refused", compiler.CompiledAmount{DynamicKind: compiler.DynamicAmountSourcePower}, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			quantity, ok := positiveQuantity(tc.amount, tc.dynamic)
			if ok != tc.ok {
				t.Fatalf("quantity=%+v ok=%v, want %v", quantity, ok, tc.ok)
			}
			if ok && tc.amount.Known && quantity.Value() != tc.amount.Value {
				t.Fatalf("fixed quantity=%d, want %d", quantity.Value(), tc.amount.Value)
			}
		})
	}
}
