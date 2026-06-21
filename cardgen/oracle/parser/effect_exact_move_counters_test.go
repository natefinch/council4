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

// TestExactMoveCountersAccepts covers the recognized counter-movement forms: a
// single named counter and the kind-agnostic "all counters" form (each from the
// source permanent onto one target creature), the distributed "any number of
// counters ... onto other creatures" group form (Forgotten Ancient), and the
// two-target forms that read counters from a first target onto a second target
// (Nesting Grounds, Daghatar, Fate Transfer).
func TestExactMoveCountersAccepts(t *testing.T) {
	t.Parallel()
	cases := []struct {
		source     string
		card       string
		all        bool
		kind       bool
		distribute bool
		fromTarget bool
		anyKind    bool
	}{
		{"Move a +1/+1 counter from this creature onto target creature.", "Steed", false, true, false, false, false},
		{"Move a +1/+1 counter from this artifact onto target creature you control.", "Weapon Rack", false, true, false, false, false},
		{"Move a +1/+1 counter from Afiya Grove onto target creature.", "Afiya Grove", false, true, false, false, false},
		{"Move all counters from this permanent onto target creature.", "The Ozolith", true, false, false, false, false},
		{"Move any number of +1/+1 counters from this creature onto other creatures.", "Forgotten Ancient", false, true, true, false, false},
		{"Move a counter from target permanent you control onto a second target permanent.", "Nesting Grounds", false, false, false, true, true},
		{"Move a +1/+1 counter from target creature onto a second target creature.", "Daghatar the Adamant", false, true, false, true, false},
		{"Move all counters from target creature onto another target creature.", "Fate Transfer", true, false, false, true, false},
		{"Move a counter from target creature an opponent controls onto target creature you control.", "Rikku, Resourceful Guardian", false, false, false, true, true},
		{"Move a +1/+1 counter from target creature you control onto another target creature you control.", "Combine Guildmage", false, true, false, true, false},
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
		if effect.MoveCountersDistribute != tc.distribute {
			t.Errorf("MoveCountersDistribute(%q) = %v, want %v", tc.source, effect.MoveCountersDistribute, tc.distribute)
		}
		if effect.MoveCountersFromTarget != tc.fromTarget {
			t.Errorf("MoveCountersFromTarget(%q) = %v, want %v", tc.source, effect.MoveCountersFromTarget, tc.fromTarget)
		}
		if effect.MoveCountersAnyKind != tc.anyKind {
			t.Errorf("MoveCountersAnyKind(%q) = %v, want %v", tc.source, effect.MoveCountersAnyKind, tc.anyKind)
		}
	}
}

// TestExactMoveCountersRejects keeps the unmodeled shapes failing closed: the
// two-target move onto a destination constrained by a relational "with the same
// controller" qualifier (Simic Guildmage) is not exactly representable.
func TestExactMoveCountersRejects(t *testing.T) {
	t.Parallel()
	cases := []struct{ source, card string }{
		{"Move a +1/+1 counter from target creature onto another target creature with the same controller.", "Simic Guildmage"},
	}
	for _, tc := range cases {
		effect := moveCountersEffect(t, tc.source, tc.card)
		if effect.Exact {
			t.Errorf("Exact(%q) = true, want false", tc.source)
		}
	}
}
