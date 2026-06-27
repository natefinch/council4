package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
)

// TestLowerKickerEntersWithCounters verifies that a "If this creature was
// kicked, it enters with N +1/+1 counters on it." clause (Kavu Aggressor's
// headline ability) lowers to a conditional EntersWithCounters replacement whose
// condition is the typed EventPermanentWasKicked predicate.
func TestLowerKickerEntersWithCounters(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Kavu Aggressor",
		Layout:     "normal",
		TypeLine:   "Creature — Kavu",
		ManaCost:   "{2}{R}",
		OracleText: "Kicker {4} (You may pay an additional {4} as you cast this spell.)\nThis creature can't block.\nIf this creature was kicked, it enters with a +1/+1 counter on it.",
	})
	if len(face.ReplacementAbilities) != 1 {
		t.Fatalf("got %d replacement abilities, want 1", len(face.ReplacementAbilities))
	}
	replacement := face.ReplacementAbilities[0].Replacement
	if len(replacement.EntersWithCounters) != 1 {
		t.Fatalf("got %d counter placements, want 1", len(replacement.EntersWithCounters))
	}
	placement := replacement.EntersWithCounters[0]
	if placement.Kind != counter.PlusOnePlusOne {
		t.Fatalf("counter kind = %v, want +1/+1", placement.Kind)
	}
	if placement.Amount != 1 {
		t.Fatalf("counter amount = %d, want 1", placement.Amount)
	}
	if !replacement.Condition.Exists {
		t.Fatal("replacement has no condition")
	}
	if !replacement.Condition.Val.EventPermanentWasKicked {
		t.Fatalf("condition is not EventPermanentWasKicked: %#v", replacement.Condition.Val)
	}
}

// TestLowerKickerEntersWithCountersNamedSource verifies the named-source phrasing
// ("If Gnarlid Colony was kicked, ...") lowers the same way as the "this
// creature" pronoun phrasing, confirming the parser recognizes a card-name
// subject as the kicked source.
func TestLowerKickerEntersWithCountersNamedSource(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Gnarlid Colony",
		Layout:     "normal",
		TypeLine:   "Creature — Beast",
		ManaCost:   "{1}{G}",
		OracleText: "Kicker {2}{G} (You may pay an additional {2}{G} as you cast this spell.)\nIf Gnarlid Colony was kicked, it enters with two +1/+1 counters on it.\nEach creature you control with a +1/+1 counter on it has trample.",
	})
	var found bool
	for _, ability := range face.ReplacementAbilities {
		replacement := ability.Replacement
		if len(replacement.EntersWithCounters) == 1 &&
			replacement.Condition.Exists &&
			replacement.Condition.Val.EventPermanentWasKicked {
			if got := replacement.EntersWithCounters[0].Amount; got != 2 {
				t.Fatalf("counter amount = %d, want 2", got)
			}
			found = true
		}
	}
	if !found {
		t.Fatal("no kicker enters-with-counters replacement lowered")
	}
}

// TestLowerKickerEntersWithCountersCombinedUnsupported verifies that the combined
// "and with <keyword>" kicker grant (Faerie Squadron) stays unsupported: the
// conditional keyword half is not representable, so lowering it as a bare
// enters-with-counters replacement would silently drop the keyword. The family
// must fail closed.
func TestLowerKickerEntersWithCountersCombinedUnsupported(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Faerie Squadron",
		Layout:     "normal",
		TypeLine:   "Creature — Faerie",
		ManaCost:   "{4}{U}",
		OracleText: "Flash\nKicker {1}{U} (You may pay an additional {1}{U} as you cast this spell.)\nFlying\nIf this creature was kicked, it enters with two +1/+1 counters on it and with flying.",
	})
}
