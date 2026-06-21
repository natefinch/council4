package parser

import "testing"

func moveCountersEffect(t *testing.T, source, cardName string) EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{CardName: cardName})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("Parse(%q) shape = %#v", source, document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectMoveCounters {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	return effects[0]
}

// TestExactMoveCountersAccepts covers the recognized single-target
// counter-movement forms: a single named counter and the kind-agnostic "all
// counters" form, each from the source permanent (a self reference) onto one
// target creature.
func TestExactMoveCountersAccepts(t *testing.T) {
	t.Parallel()
	cases := []struct {
		source string
		card   string
		all    bool
		kind   bool
	}{
		{"Move a +1/+1 counter from this creature onto target creature.", "Steed", false, true},
		{"Move a +1/+1 counter from this artifact onto target creature you control.", "Weapon Rack", false, true},
		{"Move a +1/+1 counter from Afiya Grove onto target creature.", "Afiya Grove", false, true},
		{"Move all counters from this permanent onto target creature.", "The Ozolith", true, false},
	}
	for _, tc := range cases {
		effect := moveCountersEffect(t, tc.source, tc.card)
		if !effect.Exact {
			t.Errorf("Exact(%q) = false, want true", tc.source)
		}
		if effect.MoveCountersAll != tc.all {
			t.Errorf("MoveCountersAll(%q) = %v, want %v", tc.source, effect.MoveCountersAll, tc.all)
		}
		if effect.CounterKnown != tc.kind {
			t.Errorf("CounterKnown(%q) = %v, want %v", tc.source, effect.CounterKnown, tc.kind)
		}
	}
}

// TestExactMoveCountersRejects keeps the unmodeled shapes failing closed: the
// group-distribution form ("onto other creatures", no target — Forgotten
// Ancient) and the from-another-target form ("from target creature", which is
// not a self source — Fate Transfer) are not recognized as exact.
func TestExactMoveCountersRejects(t *testing.T) {
	t.Parallel()
	cases := []struct{ source, card string }{
		{"Move any number of +1/+1 counters from this creature onto other creatures.", "Forgotten Ancient"},
		{"Move all counters from target creature onto another target creature.", "Fate Transfer"},
	}
	for _, tc := range cases {
		effect := moveCountersEffect(t, tc.source, tc.card)
		if effect.Exact {
			t.Errorf("Exact(%q) = true, want false", tc.source)
		}
	}
}
