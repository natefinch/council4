package parser

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// TestParseBlockingCreaturesPumpAmount checks the "+N/+N for each creature
// blocking it" pump amount types to the combat-state blocking-creatures count,
// carries the "+N/+N" multiplier, and records the "it" referent span so the
// lowerer can bind the count to the pumped permanent.
func TestParseBlockingCreaturesPumpAmount(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		text       string
		multiplier int
	}{
		{"self", "Whenever this creature becomes blocked, it gets +2/+2 until end of turn for each creature blocking it.", 2},
		{"other creature", "Whenever a Beast becomes blocked, it gets +1/+1 until end of turn for each creature blocking it.", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(tc.text, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
				t.Fatalf("abilities = %#v", document.Abilities)
			}
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v", effects)
			}
			amount := effects[0].Amount
			if amount.DynamicKind != EffectDynamicAmountCreaturesBlockingSource ||
				amount.DynamicForm != EffectDynamicAmountFormForEach ||
				amount.Multiplier != tc.multiplier ||
				amount.ReferenceSpan == (shared.Span{}) {
				t.Fatalf("amount = %#v", amount)
			}
		})
	}
}

// TestParseBlockingCreaturesBeyondFirstUnrecognized keeps the "beyond the first"
// (negative Rampage) free-text variant out of the all-blockers count so it stays
// fail-closed rather than counting one blocker too many.
func TestParseBlockingCreaturesBeyondFirstUnrecognized(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Whenever this creature becomes blocked, it gets -1/-1 until end of turn for each creature blocking it beyond the first.",
		Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 {
		t.Fatalf("effects = %#v", effects)
	}
	if effects[0].Amount.DynamicKind == EffectDynamicAmountCreaturesBlockingSource {
		t.Fatalf("beyond-the-first variant must not type as the all-blockers count: %#v", effects[0].Amount)
	}
}
