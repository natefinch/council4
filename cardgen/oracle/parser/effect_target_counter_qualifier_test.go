package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
)

func TestParseCounterQualifiedTargetsAreExact(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		source     string
		required   bool
		kind       counter.Kind
		any        bool
		absent     bool
		kindAbsent bool
		wantMin    int
		wantMax    int
	}{
		{
			name:     "named counter",
			source:   "Destroy target creature with a +1/+1 counter on it.",
			required: true,
			kind:     counter.PlusOnePlusOne,
			wantMin:  1,
			wantMax:  1,
		},
		{
			name:     "any counter",
			source:   "Destroy target creature with a counter on it.",
			required: true,
			any:      true,
			wantMin:  1,
			wantMax:  1,
		},
		{
			name:    "no counters",
			source:  "Destroy target creature with no counters on it.",
			absent:  true,
			wantMin: 1,
			wantMax: 1,
		},
		{
			name:       "excluded counter",
			source:     "Destroy target creature without a +1/+1 counter on it.",
			kind:       counter.PlusOnePlusOne,
			kindAbsent: true,
			wantMin:    1,
			wantMax:    1,
		},
		{
			name:     "plural named counter",
			source:   "Destroy up to two target creatures with +1/+1 counters on them.",
			required: true,
			kind:     counter.PlusOnePlusOne,
			wantMin:  0,
			wantMax:  2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			targets := firstDestroyTargets(t, test.source)
			if len(targets) != 1 {
				t.Fatalf("targets = %d, want 1", len(targets))
			}
			target := targets[0]
			if !target.Exact {
				t.Fatalf("target = %#v, want exact", target)
			}
			if target.Cardinality.Min != test.wantMin || target.Cardinality.Max != test.wantMax {
				t.Fatalf("cardinality = [%d,%d], want [%d,%d]", target.Cardinality.Min, target.Cardinality.Max, test.wantMin, test.wantMax)
			}
			selection := target.Selection
			if selection.CounterRequired != test.required ||
				selection.CounterKind != test.kind ||
				selection.CounterAny != test.any ||
				selection.CounterAbsent != test.absent ||
				selection.CounterKindAbsent != test.kindAbsent {
				t.Fatalf("selection = %#v", selection)
			}
		})
	}
}

func TestCounterQualifiedTargetPronounIsNotAFreeReference(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Target creature with a +1/+1 counter on it gains trample until end of turn.",
		Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %#v", document.Abilities)
	}
	if len(document.Abilities[0].SemanticReferences) != 0 {
		t.Fatalf("semantic references = %#v, want none", document.Abilities[0].SemanticReferences)
	}
}
