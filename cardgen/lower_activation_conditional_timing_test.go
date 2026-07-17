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

// TestLowerActivatedEventHistoryConditionWithSorceryTimingTail proves an
// activated ability whose gate conjoins an event-history attack-count condition
// with a sorcery-timing tail ("Activate only if you attacked with three or more
// creatures this turn and only as a sorcery.") lowers both the event-history
// activation condition and the sorcery timing restriction. This is the shape of
// Temple of Civilization's transform activation (Ojer Taq, Deepest Foundation);
// previously the conjoined "and only as a sorcery" tail sat inside the
// event-history clause, so the "this turn" window suffix was no longer the
// clause's trailing words and the whole gate failed to lower.
func TestLowerActivatedEventHistoryConditionWithSorceryTimingTail(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Vanguard",
		Layout:     "normal",
		TypeLine:   "Creature — Soldier",
		OracleText: "{2}{W}: Draw a card. Activate only if you attacked with three or more creatures this turn and only as a sorcery.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	ability := face.ActivatedAbilities[0]
	if ability.Timing != game.SorceryOnly {
		t.Fatalf("timing = %v, want SorceryOnly", ability.Timing)
	}
	if !ability.ActivationCondition.Exists || !ability.ActivationCondition.Val.EventHistory.Exists {
		t.Fatalf("activation condition = %#v, want event-history gate", ability.ActivationCondition)
	}
	eventHistory := ability.ActivationCondition.Val.EventHistory.Val
	if eventHistory.MinCount != 3 {
		t.Fatalf("event-history MinCount = %d, want 3", eventHistory.MinCount)
	}
	if eventHistory.Window != game.EventHistoryCurrentTurn {
		t.Fatalf("event-history window = %v, want current turn", eventHistory.Window)
	}
	if eventHistory.Pattern.Event != game.EventAttackerDeclared ||
		eventHistory.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("event-history pattern = %#v, want your attacker-declared events", eventHistory.Pattern)
	}
}
