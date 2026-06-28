package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
)

// TestParseRemoveCounterFromSelfKindUnspecified covers the activation cost
// "Remove a/N counter(s) from this <permanent>" with no named counter kind, as
// printed on the -1/-1 counter creatures (Loch Mare, Moonlit Lamenter, ...). The
// removal is recognized as a self-source counter cost with the kind left
// unspecified, so lowering can emit an any-kind removal the payer resolves
// against whatever counters the source carries.
func TestParseRemoveCounterFromSelfKindUnspecified(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		source string
		amount int
	}{
		{"a counter from this creature", "Remove a counter from this creature: Draw a card.", 1},
		{"two counters from this creature", "Remove two counters from this creature: Draw a card.", 2},
		{"a counter from this permanent", "Remove a counter from this permanent: Draw a card.", 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := soleCostComponent(t, test.source)
			if component.Kind != CostComponentRemoveCounter {
				t.Fatalf("kind = %v, want remove counter", component.Kind)
			}
			if !component.SourceSelf {
				t.Fatalf("component = %#v, want source-self", component)
			}
			if component.RemoveCounterAmong {
				t.Fatalf("component = %#v, want not among removal", component)
			}
			if !component.AmountKnown || component.AmountValue != test.amount {
				t.Fatalf("amount = (%d, known %t), want %d", component.AmountValue, component.AmountKnown, test.amount)
			}
			if component.CounterKindKnown {
				t.Fatalf("component = %#v, want counter kind unspecified", component)
			}
		})
	}
}

// TestParseRemoveCounterFromSelfNamedKind confirms that a named counter kind is
// still recognized on the self-source removal cost and reported with the exact
// kind, so the kind-unspecified relaxation does not erase named-kind precision.
func TestParseRemoveCounterFromSelfNamedKind(t *testing.T) {
	t.Parallel()
	component := soleCostComponent(t, "Remove a charge counter from this artifact: Draw a card.")
	if component.Kind != CostComponentRemoveCounter {
		t.Fatalf("kind = %v, want remove counter", component.Kind)
	}
	if !component.SourceSelf {
		t.Fatalf("component = %#v, want source-self", component)
	}
	if !component.CounterKindKnown || component.CounterKind != counter.Charge {
		t.Fatalf("counter = (%v, known %t), want charge", component.CounterKind, component.CounterKindKnown)
	}
}
