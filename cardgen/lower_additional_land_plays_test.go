package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerAdditionalLandPlaysOneShot proves that the spell clause "You may play
// an additional land this turn." lowers to an ApplyRule primitive carrying the
// controller-scoped RuleEffectAdditionalLandPlays for the turn.
func TestLowerAdditionalLandPlaysOneShot(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Explore",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Draw a card. You may play an additional land this turn.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("expected a spell ability")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want two primitives", mode.Sequence)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.Draw); !ok {
		t.Fatalf("first primitive = %T, want game.Draw", mode.Sequence[0].Primitive)
	}
	apply, ok := mode.Sequence[1].Primitive.(game.ApplyRule)
	if !ok {
		t.Fatalf("second primitive = %T, want game.ApplyRule", mode.Sequence[1].Primitive)
	}
	if apply.Duration != game.DurationThisTurn {
		t.Fatalf("duration = %v, want DurationThisTurn", apply.Duration)
	}
	if len(apply.RuleEffects) != 1 {
		t.Fatalf("rule effects = %#v, want one", apply.RuleEffects)
	}
	effect := apply.RuleEffects[0]
	if effect.Kind != game.RuleEffectAdditionalLandPlays {
		t.Fatalf("kind = %v, want RuleEffectAdditionalLandPlays", effect.Kind)
	}
	if effect.AffectedPlayer != game.PlayerYou {
		t.Fatalf("affected player = %v, want PlayerYou", effect.AffectedPlayer)
	}
	if effect.AdditionalLandPlays != 1 {
		t.Fatalf("additional land plays = %d, want 1", effect.AdditionalLandPlays)
	}
}

// TestLowerAdditionalLandPlaysOneShotMultiple proves the multi-land "up to N"
// wording carries the parsed count.
func TestLowerAdditionalLandPlaysOneShotMultiple(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Summer Bloom",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "You may play up to three additional lands this turn.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyRule)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyRule", mode.Sequence[0].Primitive)
	}
	if got := apply.RuleEffects[0].AdditionalLandPlays; got != 3 {
		t.Fatalf("additional land plays = %d, want 3", got)
	}
}

// TestLowerAdditionalLandPlaysStatic proves that the static "on each of your
// turns" wording lowers to a static ability carrying the controller-scoped
// RuleEffectAdditionalLandPlays with no duration (continuous).
func TestLowerAdditionalLandPlaysStatic(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Exploration",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "You may play an additional land on each of your turns.",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	effects := face.StaticAbilities[0].Body.RuleEffects
	if len(effects) != 1 {
		t.Fatalf("rule effects = %#v, want one", effects)
	}
	effect := effects[0]
	if effect.Kind != game.RuleEffectAdditionalLandPlays {
		t.Fatalf("kind = %v, want RuleEffectAdditionalLandPlays", effect.Kind)
	}
	if effect.AffectedPlayer != game.PlayerYou {
		t.Fatalf("affected player = %v, want PlayerYou", effect.AffectedPlayer)
	}
	if effect.AdditionalLandPlays != 1 {
		t.Fatalf("additional land plays = %d, want 1", effect.AdditionalLandPlays)
	}
	if effect.AffectedSource || effect.AffectedAttached {
		t.Fatalf("rule effect must be player-scoped: %#v", effect)
	}
}

// TestLowerAdditionalLandPlaysStaticMultiple proves the static multi-land wording
// (Azusa) carries the parsed count.
func TestLowerAdditionalLandPlaysStaticMultiple(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Azusa",
		Layout:     "normal",
		TypeLine:   "Creature",
		OracleText: "You may play two additional lands on each of your turns.",
	})
	if got := face.StaticAbilities[0].Body.RuleEffects[0].AdditionalLandPlays; got != 2 {
		t.Fatalf("additional land plays = %d, want 2", got)
	}
}
