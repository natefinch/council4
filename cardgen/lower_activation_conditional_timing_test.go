package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerActivatedConditionWithSorceryTimingTail proves an activated ability
// whose gate conjoins a state condition with a sorcery-timing tail ("Activate
// only if you control an enchantment and only as a sorcery.") lowers both the
// activation condition and the sorcery timing restriction. Previously the
// conjoined "and only as a sorcery" tail blocked the whole gate as an
// unsupported activation condition.
func TestLowerActivatedConditionWithSorceryTimingTail(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Geist",
		Layout:     "normal",
		TypeLine:   "Creature — Spirit",
		OracleText: "{2}{W}: Draw a card. Activate only if you control an enchantment and only as a sorcery.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	ability := face.ActivatedAbilities[0]
	if ability.Timing != game.SorceryOnly {
		t.Fatalf("timing = %v, want SorceryOnly", ability.Timing)
	}
	if !ability.ActivationCondition.Exists || !ability.ActivationCondition.Val.ControlsMatching.Exists {
		t.Fatalf("activation condition = %#v, want controls-matching gate", ability.ActivationCondition)
	}
}

// TestLowerActivatedConditionWithFrequencyTimingTail proves the once-per-turn
// timing tail conjoined to a state condition lowers to OncePerTurn.
func TestLowerActivatedConditionWithFrequencyTimingTail(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Corruptor",
		Layout:     "normal",
		TypeLine:   "Creature — Insect",
		OracleText: "{B}: Draw a card. Activate only if an opponent has three or more poison counters and only once each turn.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	ability := face.ActivatedAbilities[0]
	if ability.Timing != game.OncePerTurn {
		t.Fatalf("timing = %v, want OncePerTurn", ability.Timing)
	}
	if !ability.ActivationCondition.Exists || ability.ActivationCondition.Val.AnyOpponentPoisonAtLeast != 3 {
		t.Fatalf("activation condition = %#v, want opponent poison >= 3", ability.ActivationCondition)
	}
}

// TestLowerActivatedConditionWithPlayerTurnTimingTail proves the "during your
// turn" timing tail conjoined to a state condition lowers to DuringYourTurn.
func TestLowerActivatedConditionWithPlayerTurnTimingTail(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Defender",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "{2}{W}: Draw a card. Activate only if you control no creatures and only during your turn.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	ability := face.ActivatedAbilities[0]
	if ability.Timing != game.DuringYourTurn {
		t.Fatalf("timing = %v, want DuringYourTurn", ability.Timing)
	}
	if !ability.ActivationCondition.Exists || !ability.ActivationCondition.Val.ControlsMatching.Exists {
		t.Fatalf("activation condition = %#v, want controls-matching gate", ability.ActivationCondition)
	}
}
