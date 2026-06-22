package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

func TestLowerActivatedAbilitySourceCombatStateCondition(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Raider",
		Layout:     "normal",
		TypeLine:   "Creature — Test",
		OracleText: "{1}: Target creature gets +2/+2 until end of turn. Activate only if this creature is attacking.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	ability := face.ActivatedAbilities[0]
	if !ability.ActivationCondition.Exists || !ability.ActivationCondition.Val.ObjectMatches.Exists {
		t.Fatalf("activation condition = %#v, want source object match", ability.ActivationCondition)
	}
	if got := ability.ActivationCondition.Val.ObjectMatches.Val.CombatState; got != game.CombatStateAttacking {
		t.Fatalf("combat state = %v, want attacking", got)
	}
}

func TestLowerActivatedAbilitySourcePowerCondition(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bruiser",
		Layout:     "normal",
		TypeLine:   "Creature — Test",
		OracleText: "{1}: Target creature gets +2/+2 until end of turn. Activate only if this creature's power is 4 or greater.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	condition := face.ActivatedAbilities[0].ActivationCondition
	if !condition.Exists || !condition.Val.ObjectMatches.Exists || !condition.Val.ObjectMatches.Val.Power.Exists {
		t.Fatalf("activation condition = %#v, want source power match", condition)
	}
	if got := condition.Val.ObjectMatches.Val.Power.Val.Value; got != 4 {
		t.Fatalf("power threshold = %d, want 4", got)
	}
}

func TestLowerActivatedAbilityOpponentPoisonCondition(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Corruptor",
		Layout:     "normal",
		TypeLine:   "Creature — Test",
		OracleText: "{B}: Draw a card. Activate only if an opponent has three or more poison counters.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	condition := face.ActivatedAbilities[0].ActivationCondition
	if !condition.Exists || condition.Val.AnyOpponentPoisonAtLeast != 3 {
		t.Fatalf("activation condition = %#v, want opponent poison >= 3", condition)
	}
}

func TestLowerActivatedAbilityLifeAboveStartingCondition(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Pilgrim",
		Layout:     "normal",
		TypeLine:   "Creature — Test",
		OracleText: "{1}: Draw a card. Activate only if you have at least 10 life more than your starting life total.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	condition := face.ActivatedAbilities[0].ActivationCondition
	if !condition.Exists || condition.Val.ControllerLifeAtLeastAboveStarting != 10 {
		t.Fatalf("activation condition = %#v, want life-above-starting >= 10", condition)
	}
}

func TestLowerActivatedAbilityLifeAtMostCondition(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Ascetic",
		Layout:     "normal",
		TypeLine:   "Creature — Test",
		OracleText: "{1}: Draw a card. Activate only if you have 5 or less life.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	condition := face.ActivatedAbilities[0].ActivationCondition
	if !condition.Exists || !condition.Val.ControllerLifeAtMost.Exists || condition.Val.ControllerLifeAtMost.Val != 5 {
		t.Fatalf("activation condition = %#v, want life at most 5", condition)
	}
}

func TestLowerActivatedAbilityCreatedTokenThisTurnCondition(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Idol",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{T}: Draw a card. Activate only if you created a token this turn.",
	})
	condition := face.ActivatedAbilities[0].ActivationCondition
	if !condition.Exists || !condition.Val.ControllerCreatedTokenThisTurn {
		t.Fatalf("activation condition = %#v, want created-token-this-turn", condition)
	}
}

func TestLowerActivatedAbilityExactHandSizeCondition(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Librarian",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{1}: Draw a card. Activate only if you have exactly seven cards in hand.",
	})
	condition := face.ActivatedAbilities[0].ActivationCondition
	if !condition.Exists || !condition.Val.ControllerHandSizeExactly.Exists ||
		condition.Val.ControllerHandSizeExactly.Val != 7 {
		t.Fatalf("activation condition = %#v, want exactly seven cards in hand", condition)
	}
}

func TestLowerActivatedAbilityControlsKeywordCondition(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Tapper",
		Layout:     "normal",
		TypeLine:   "Creature — Test",
		OracleText: "{1}: Draw a card. Activate only if you control a creature with flying.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	condition := face.ActivatedAbilities[0].ActivationCondition
	if !condition.Exists || !condition.Val.ControlsMatching.Exists {
		t.Fatalf("activation condition = %#v, want controls-matching selection", condition)
	}
	if got := condition.Val.ControlsMatching.Val.Selection.Keyword; got != game.Flying {
		t.Fatalf("selection keyword = %v, want flying", got)
	}
}

// TestLowerActivatedAbilityUnsupportedCombatStateFailsClosed verifies a combat
// involvement the runtime can't model ("blocked") leaves the ability blocked on
// an unsupported activation condition rather than guessing a meaning.
func TestLowerActivatedAbilityUnsupportedCombatStateFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Skulker",
		Layout:     "normal",
		TypeLine:   "Creature — Test",
		OracleText: "{1}: Draw a card. Activate only if this creature is blocked.",
		Power:      new("2"),
		Toughness:  new("2"),
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(diagnostics, func(diagnostic shared.Diagnostic) bool {
		return diagnostic.Summary == "unsupported activation condition"
	}) {
		t.Fatalf("diagnostics = %#v, want unsupported activation condition for blocked state", diagnostics)
	}
}
