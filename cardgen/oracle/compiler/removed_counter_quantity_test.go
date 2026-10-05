package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestCompileRemovedCounterQuantityIdentity(t *testing.T) {
	t.Parallel()
	amount := parser.EffectAmountSyntax{
		DynamicKind: parser.EffectDynamicAmountRemovedCounterCount, ProducerClauseID: 37, Multiplier: 2, Addend: 1,
		Text: "diagnostic-only", Span: shared.Span{Start: shared.Position{Offset: 9000}},
	}
	got := compileTypedAmount(amount)
	if got.DynamicKind != DynamicAmountRemovedCounterCount || got.ProducerClauseID != 37 ||
		got.Multiplier != 2 || got.Addend != 1 {
		t.Fatalf("compiled=%#v", got)
	}
	amount.ProducerClauseID = 0
	if got := compileTypedAmount(amount); got.ProducerClauseID != 0 {
		t.Fatalf("missing identity was invented: %#v", got)
	}
}
