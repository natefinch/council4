package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerBecameTargetFirstTimeEachTurnCapsTrigger covers the Valiant ability
// word (Bloomburrow) and the Glasskite spirits: a became-target trigger
// restricted to "the first time each turn" lowers to the existing
// object-became-target event with a once-per-turn cap.
func TestLowerBecameTargetFirstTimeEachTurnCapsTrigger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		text      string
		wantCause game.TriggerControllerFilter
		wantMax   int
	}{
		{
			name:      "valiant ability word you control",
			text:      "Valiant — Whenever this creature becomes the target of a spell or ability you control for the first time each turn, put a +1/+1 counter on it.",
			wantCause: game.TriggerControllerYou,
			wantMax:   1,
		},
		{
			name:      "glasskite any controller counter",
			text:      "Whenever this creature becomes the target of a spell or ability for the first time each turn, draw a card.",
			wantCause: game.TriggerControllerAny,
			wantMax:   1,
		},
		{
			name:      "no first-time qualifier leaves trigger uncapped",
			text:      "Whenever this creature becomes the target of a spell or ability you control, put a +1/+1 counter on it.",
			wantCause: game.TriggerControllerYou,
			wantMax:   0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Mouse",
				Layout:     "normal",
				TypeLine:   "Creature — Mouse Soldier",
				OracleText: test.text,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
			}
			ta := face.TriggeredAbilities[0]
			if ta.Trigger.Pattern.Event != game.EventObjectBecameTarget {
				t.Errorf("event = %v, want EventObjectBecameTarget", ta.Trigger.Pattern.Event)
			}
			if ta.Trigger.Pattern.Source != game.TriggerSourceSelf {
				t.Errorf("source = %v, want TriggerSourceSelf", ta.Trigger.Pattern.Source)
			}
			if ta.Trigger.Pattern.CauseController != test.wantCause {
				t.Errorf("cause controller = %v, want %v", ta.Trigger.Pattern.CauseController, test.wantCause)
			}
			if ta.MaxTriggersPerTurn != test.wantMax {
				t.Errorf("MaxTriggersPerTurn = %d, want %d", ta.MaxTriggersPerTurn, test.wantMax)
			}
		})
	}
}
