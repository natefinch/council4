package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerCastAsThoughFlashOneShot proves that "You may cast spells this turn as
// though they had flash." lowers to an ApplyRule carrying the controller-scoped
// RuleEffectCastSpellsAsThoughFlash for the turn (Borne Upon a Wind).
func TestLowerCastAsThoughFlashOneShot(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Borne Upon a Wind",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "You may cast spells this turn as though they had flash.\nDraw a card.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("expected a spell ability")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want two primitives", mode.Sequence)
	}
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyRule)
	if !ok {
		t.Fatalf("first primitive = %T, want game.ApplyRule", mode.Sequence[0].Primitive)
	}
	if apply.Duration != game.DurationThisTurn {
		t.Fatalf("duration = %v, want DurationThisTurn", apply.Duration)
	}
	if len(apply.RuleEffects) != 1 {
		t.Fatalf("rule effects = %#v, want one", apply.RuleEffects)
	}
	effect := apply.RuleEffects[0]
	if effect.Kind != game.RuleEffectCastSpellsAsThoughFlash {
		t.Fatalf("kind = %v, want RuleEffectCastSpellsAsThoughFlash", effect.Kind)
	}
	if effect.AffectedPlayer != game.PlayerYou {
		t.Fatalf("affected player = %v, want PlayerYou", effect.AffectedPlayer)
	}
	if _, ok := mode.Sequence[1].Primitive.(game.Draw); !ok {
		t.Fatalf("second primitive = %T, want game.Draw", mode.Sequence[1].Primitive)
	}
}

// TestLowerCastAsThoughFlashInActivatedAbility proves the same permission lowers
// inside an activated ability body (Emergence Zone).
func TestLowerCastAsThoughFlashInActivatedAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Emergence Zone",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {C}.\n{1}, {T}, Sacrifice this land: You may cast spells this turn as though they had flash.",
	})
	var found bool
	for _, ability := range face.ActivatedAbilities {
		for _, ins := range ability.Content.Modes[0].Sequence {
			apply, ok := ins.Primitive.(game.ApplyRule)
			if ok && len(apply.RuleEffects) == 1 &&
				apply.RuleEffects[0].Kind == game.RuleEffectCastSpellsAsThoughFlash {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("expected an activated ability granting RuleEffectCastSpellsAsThoughFlash")
	}
}
