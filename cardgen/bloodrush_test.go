package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerBloodrushKeyword verifies that the Bloodrush ability word lowers as a
// rules-free label: the activated ability functions from the hand, discards
// itself as an additional cost, and keeps its printed mana cost. Bloodrush adds
// no rules of its own (CR uses an ability word purely for flavor), so the body
// is the entire activated ability.
func TestLowerBloodrushKeyword(t *testing.T) {
	t.Parallel()
	power := "3"
	toughness := "3"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bloodrusher",
		Layout:     "normal",
		TypeLine:   "Creature — Beast",
		ManaCost:   "{1}{R}",
		OracleText: "Bloodrush — {R}, Discard this card: Target attacking creature gets +3/+3 until end of turn.",
		Power:      &power,
		Toughness:  &toughness,
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if ability.ZoneOfFunction != zone.Hand {
		t.Fatalf("zone of function = %v, want Hand", ability.ZoneOfFunction)
	}
	if !ability.ManaCost.Exists {
		t.Fatal("expected a printed mana cost on the Bloodrush activation")
	}
	if len(ability.AdditionalCosts) != 1 {
		t.Fatalf("additional costs = %d, want 1", len(ability.AdditionalCosts))
	}
	discard := ability.AdditionalCosts[0]
	if discard.Kind != cost.AdditionalDiscard {
		t.Fatalf("additional cost kind = %v, want AdditionalDiscard", discard.Kind)
	}
	if discard.Source != zone.Hand {
		t.Fatalf("discard source = %v, want Hand", discard.Source)
	}
}

// TestLowerBloodrushRejectsUnrelatedAbilityWord guards the rules-free gate: an
// unknown activated ability word still fails closed so only the curated set of
// flavor labels (which now includes Bloodrush) lowers.
func TestLowerBloodrushRejectsUnrelatedAbilityWord(t *testing.T) {
	t.Parallel()
	power := "3"
	toughness := "3"
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Unknown Word",
		Layout:     "normal",
		TypeLine:   "Creature — Beast",
		ManaCost:   "{1}{R}",
		OracleText: "Frothburst — {R}, Discard this card: Target attacking creature gets +3/+3 until end of turn.",
		Power:      &power,
		Toughness:  &toughness,
	})
}
