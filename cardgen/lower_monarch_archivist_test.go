package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerArchivistOfGondorTriggers locks in the two new trigger recognizers
// behind Archivist of Gondor: "your commander" as a combat-damage trigger
// subject (DamageSourceSelection.MatchCommander) and "the monarch's end step" as
// a monarch-scoped beginning-of-step trigger (TriggerPlayerMonarch).
func TestLowerArchivistOfGondorTriggers(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Archivist of Gondor",
		Layout:   "normal",
		TypeLine: "Creature — Human Advisor",
		OracleText: "When your commander deals combat damage to a player, if there is no monarch, you become the monarch.\n" +
			"At the beginning of the monarch's end step, that player draws a card.",
		Power:     new("2"),
		Toughness: new("3"),
	})
	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("got %d triggered abilities, want 2", len(face.TriggeredAbilities))
	}

	// Gap 1: commander combat damage becomes monarch when there is no monarch.
	commanderTrigger := face.TriggeredAbilities[0].Trigger
	if commanderTrigger.Pattern.Event != game.EventDamageDealt {
		t.Fatalf("trigger 0 event = %v, want EventDamageDealt", commanderTrigger.Pattern.Event)
	}
	if commanderTrigger.Pattern.Subject != game.TriggerSubjectDamageSource {
		t.Fatalf("trigger 0 subject = %v, want TriggerSubjectDamageSource", commanderTrigger.Pattern.Subject)
	}
	if !commanderTrigger.Pattern.DamageSourceSelection.MatchCommander {
		t.Fatal("trigger 0 DamageSourceSelection.MatchCommander = false, want true")
	}
	if !commanderTrigger.Pattern.RequireCombatDamage {
		t.Fatal("trigger 0 RequireCombatDamage = false, want true")
	}
	if commanderTrigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("trigger 0 controller = %v, want TriggerControllerYou", commanderTrigger.Pattern.Controller)
	}
	if !commanderTrigger.InterveningCondition.Exists ||
		!commanderTrigger.InterveningCondition.Val.NoMonarch {
		t.Fatal("trigger 0 intervening condition = _, want NoMonarch")
	}

	// Gap 2: at the beginning of the monarch's end step, that player draws.
	endStepTrigger := face.TriggeredAbilities[1].Trigger
	if endStepTrigger.Pattern.Event != game.EventBeginningOfStep {
		t.Fatalf("trigger 1 event = %v, want EventBeginningOfStep", endStepTrigger.Pattern.Event)
	}
	if endStepTrigger.Pattern.Step != game.StepEnd {
		t.Fatalf("trigger 1 step = %v, want StepEnd", endStepTrigger.Pattern.Step)
	}
	if endStepTrigger.Pattern.Player != game.TriggerPlayerMonarch {
		t.Fatalf("trigger 1 player = %v, want TriggerPlayerMonarch", endStepTrigger.Pattern.Player)
	}
}
