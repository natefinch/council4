package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerSanguineBondTargetDrain proves "Whenever you gain life, target
// opponent loses that much life." lowers to a life-gain trigger whose body loses
// the triggering life amount (DynamicAmountEventLifeChange) from a target
// opponent.
func TestLowerSanguineBondTargetDrain(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Sanguine Bond",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		ManaCost:   "{3}{B}{B}",
		OracleText: "Whenever you gain life, target opponent loses that much life.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ability := face.TriggeredAbilities[0]
	if ability.Trigger.Pattern.Event != game.EventLifeGained {
		t.Fatalf("trigger event = %v, want EventLifeGained", ability.Trigger.Pattern.Event)
	}
	prim := ability.Content.Modes[0].Sequence[0].Primitive
	lose, ok := prim.(game.LoseLife)
	if !ok {
		t.Fatalf("primitive = %T, want game.LoseLife", prim)
	}
	if dyn := lose.Amount.DynamicAmount(); !dyn.Exists || dyn.Val.Kind != game.DynamicAmountEventLifeChange {
		t.Fatalf("amount = %#v, want DynamicAmountEventLifeChange", lose.Amount)
	}
	if lose.Player != game.TargetPlayerReference(0) {
		t.Fatalf("player = %#v, want target player 0", lose.Player)
	}
}

// TestLowerEachOpponentLifeDrain proves the group form "Whenever you gain life,
// each opponent loses that much life." lowers to a group LoseLife on opponents
// (Marauding Blight-Priest, Epicure of Blood).
func TestLowerEachOpponentLifeDrain(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Blight Priest",
		Layout:     "normal",
		TypeLine:   "Creature",
		ManaCost:   "{2}{B}",
		OracleText: "Whenever you gain life, each opponent loses that much life.",
	})
	prim := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive
	lose, ok := prim.(game.LoseLife)
	if !ok {
		t.Fatalf("primitive = %T, want game.LoseLife", prim)
	}
	if dyn := lose.Amount.DynamicAmount(); !dyn.Exists || dyn.Val.Kind != game.DynamicAmountEventLifeChange {
		t.Fatalf("amount = %#v, want DynamicAmountEventLifeChange", lose.Amount)
	}
	if lose.PlayerGroup != game.OpponentsReference() {
		t.Fatalf("group = %#v, want opponents", lose.PlayerGroup)
	}
}

// TestLowerExquisiteBloodMirror proves the mirror "Whenever an opponent loses
// life, you gain that much life." lowers to a life-loss trigger whose body gains
// the triggering life amount for the controller.
func TestLowerExquisiteBloodMirror(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Exquisite Blood",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		ManaCost:   "{4}{B}{B}",
		OracleText: "Whenever an opponent loses life, you gain that much life.",
	})
	ability := face.TriggeredAbilities[0]
	if ability.Trigger.Pattern.Event != game.EventLifeLost {
		t.Fatalf("trigger event = %v, want EventLifeLost", ability.Trigger.Pattern.Event)
	}
	prim := ability.Content.Modes[0].Sequence[0].Primitive
	gain, ok := prim.(game.GainLife)
	if !ok {
		t.Fatalf("primitive = %T, want game.GainLife", prim)
	}
	if dyn := gain.Amount.DynamicAmount(); !dyn.Exists || dyn.Val.Kind != game.DynamicAmountEventLifeChange {
		t.Fatalf("amount = %#v, want DynamicAmountEventLifeChange", gain.Amount)
	}
	if gain.Player != game.ControllerReference() {
		t.Fatalf("player = %#v, want controller", gain.Player)
	}
}

// TestLowerThatMuchLifeFailsClosedWithoutLifeTrigger proves "that much life" has
// no triggering quantity outside a life-change trigger and fails closed.
func TestLowerThatMuchLifeFailsClosedWithoutLifeTrigger(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Bare Drain",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{B}",
		OracleText: "Target opponent loses that much life.",
	})
}
