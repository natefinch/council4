package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerCantCastSpellsOpponents proves that "Your opponents can't cast spells
// this turn." lowers to a one-shot ApplyRule carrying RuleEffectCantCastSpells
// affecting the controller's opponents for the turn (Silence).
func TestLowerCantCastSpellsOpponents(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Silence",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{W}",
		OracleText: "Your opponents can't cast spells this turn.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("expected a spell ability")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want one primitive", mode.Sequence)
	}
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyRule)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyRule", mode.Sequence[0].Primitive)
	}
	if apply.Duration != game.DurationThisTurn {
		t.Fatalf("duration = %v, want DurationThisTurn", apply.Duration)
	}
	if len(apply.RuleEffects) != 1 {
		t.Fatalf("rule effects = %#v, want one", apply.RuleEffects)
	}
	effect := apply.RuleEffects[0]
	if effect.Kind != game.RuleEffectCantCastSpells {
		t.Fatalf("kind = %v, want RuleEffectCantCastSpells", effect.Kind)
	}
	if effect.AffectedPlayer != game.PlayerOpponent {
		t.Fatalf("affected player = %v, want PlayerOpponent", effect.AffectedPlayer)
	}
}

// TestLowerCantCastSpellsAllPlayers proves that "Players can't cast spells this
// turn." lowers to the all-players (PlayerAny) cast prohibition.
func TestLowerCantCastSpellsAllPlayers(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mass Silence",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{W}",
		OracleText: "Players can't cast spells this turn.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("expected a spell ability")
	}
	mode := face.SpellAbility.Val.Modes[0]
	apply, ok := mode.Sequence[0].Primitive.(game.ApplyRule)
	if !ok {
		t.Fatalf("primitive = %T, want game.ApplyRule", mode.Sequence[0].Primitive)
	}
	effect := apply.RuleEffects[0]
	if effect.Kind != game.RuleEffectCantCastSpells {
		t.Fatalf("kind = %v, want RuleEffectCantCastSpells", effect.Kind)
	}
	if effect.AffectedPlayer != game.PlayerAny {
		t.Fatalf("affected player = %v, want PlayerAny", effect.AffectedPlayer)
	}
}

// TestLowerCantCastSpellsTargetPlayerFailsClosed proves the targeted form "Target
// player can't cast spells this turn." is not lowered by the one-shot path; it
// has no supported lowering and yields no spell ability content.
func TestLowerCantCastSpellsTargetPlayerFailsClosed(t *testing.T) {
	t.Parallel()
	_, diags, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Orim's Chant",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{W}",
		OracleText: "Target player can't cast spells this turn.",
	}, "t")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatal("expected a diagnostic for the unsupported targeted cast prohibition")
	}
}
