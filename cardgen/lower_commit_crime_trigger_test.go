package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerCommitCrimeTriggerYou(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Crime Watcher",
		Layout:     "normal",
		ManaCost:   "{1}{R}",
		TypeLine:   "Creature — Goblin Rogue",
		OracleText: "Whenever you commit a crime, you gain 1 life.",
		Colors:     []string{"R"},
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	got := face.TriggeredAbilities[0]
	if got.Trigger.Type != game.TriggerWhenever {
		t.Errorf("Trigger.Type = %v, want TriggerWhenever", got.Trigger.Type)
	}
	if got.Trigger.Pattern.Event != game.EventCrimeCommitted {
		t.Errorf("Pattern.Event = %v, want EventCrimeCommitted", got.Trigger.Pattern.Event)
	}
	if got.Trigger.Pattern.Player != game.TriggerPlayerYou {
		t.Errorf("Pattern.Player = %v, want TriggerPlayerYou", got.Trigger.Pattern.Player)
	}
}

func TestLowerCommitCrimeTriggerOpponent(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Crime Punisher",
		Layout:     "normal",
		ManaCost:   "{1}{W}",
		TypeLine:   "Creature — Soldier",
		OracleText: "Whenever an opponent commits a crime, you gain 1 life.",
		Colors:     []string{"W"},
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	got := face.TriggeredAbilities[0]
	if got.Trigger.Pattern.Event != game.EventCrimeCommitted {
		t.Errorf("Pattern.Event = %v, want EventCrimeCommitted", got.Trigger.Pattern.Event)
	}
	if got.Trigger.Pattern.Player != game.TriggerPlayerOpponent {
		t.Errorf("Pattern.Player = %v, want TriggerPlayerOpponent", got.Trigger.Pattern.Player)
	}
}
