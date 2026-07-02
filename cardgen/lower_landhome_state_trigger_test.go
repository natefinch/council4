package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerLandhomeStateTrigger proves "When you control no <subtype>, sacrifice
// this creature." lowers to a state-triggered ability (CR 603.8) whose condition
// fires while the controller controls zero permanents of that subtype and whose
// body sacrifices the source permanent.
func TestLowerLandhomeStateTrigger(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		text    string
		subtype types.Sub
	}{
		{"islands", "When you control no Islands, sacrifice this creature.", types.Island},
		{"swamps", "When you control no Swamps, sacrifice this creature.", types.Swamp},
		{"forests", "When you control no Forests, sacrifice this creature.", types.Forest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Landhome",
				Layout:     "normal",
				TypeLine:   "Creature — Serpent",
				OracleText: tt.text,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
			}
			ability := face.TriggeredAbilities[0]
			if ability.Trigger.Type != game.TriggerState {
				t.Fatalf("trigger type = %v, want state trigger", ability.Trigger.Type)
			}
			if !ability.Trigger.State.Exists {
				t.Fatal("state trigger condition absent")
			}
			condition := ability.Trigger.State.Val.Condition
			if !condition.Exists {
				t.Fatal("state trigger board condition absent")
			}
			if !condition.Val.Negate {
				t.Fatalf("condition = %#v, want negated (\"control no\")", condition.Val)
			}
			controls := condition.Val.ControlsMatching
			if !controls.Exists || controls.Val.MinCount != 1 {
				t.Fatalf("controls = %#v, want at-least-one match", controls)
			}
			found := false
			for _, sub := range controls.Val.Selection.SubtypesAny {
				if sub == tt.subtype {
					found = true
				}
			}
			if !found {
				t.Fatalf("selection = %#v, want %q", controls.Val.Selection, tt.subtype)
			}
			sequence := ability.Content.Modes[0].Sequence
			if len(sequence) != 1 {
				t.Fatalf("sequence length = %d, want 1", len(sequence))
			}
			sacrifice, ok := sequence[0].Primitive.(game.Sacrifice)
			if !ok || sacrifice.Object != game.SourceCardPermanentReference() {
				t.Fatalf("primitive = %#v, want source-card sacrifice", sequence[0].Primitive)
			}
		})
	}
}
