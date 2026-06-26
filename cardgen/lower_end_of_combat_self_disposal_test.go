package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerEndOfCombatSelfDisposal proves the fragile attacker/blocker idiom
// "When this creature attacks or blocks, <self-disposal> at end of combat"
// lowers each disposal verb to its delayed end-of-combat primitive acting on
// the source permanent. Sacrifice and destroy dispose the source directly,
// "return it to its owner's hand" bounces the still-on-battlefield permanent,
// and "put it on top of its owner's library" tucks it.
func TestLowerEndOfCombatSelfDisposal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		text  string
		check func(t *testing.T, primitive game.Primitive)
	}{
		{
			name: "sacrifice",
			text: "When this creature attacks or blocks, sacrifice it at end of combat.",
			check: func(t *testing.T, primitive game.Primitive) {
				t.Helper()
				sacrifice, ok := primitive.(game.Sacrifice)
				if !ok || sacrifice.Object != game.SourceCardPermanentReference() {
					t.Fatalf("primitive = %#v, want source-card sacrifice", primitive)
				}
			},
		},
		{
			name: "destroy",
			text: "When this creature attacks or blocks, destroy it at end of combat.",
			check: func(t *testing.T, primitive game.Primitive) {
				t.Helper()
				destroy, ok := primitive.(game.Destroy)
				if !ok || destroy.Object != game.SourceCardPermanentReference() {
					t.Fatalf("primitive = %#v, want source-card destroy", primitive)
				}
			},
		},
		{
			name: "return to hand bounces",
			text: "When this creature attacks or blocks, return it to its owner's hand at end of combat.",
			check: func(t *testing.T, primitive game.Primitive) {
				t.Helper()
				bounce, ok := primitive.(game.Bounce)
				if !ok || bounce.Object != game.SourceCardPermanentReference() {
					t.Fatalf("primitive = %#v, want source-card bounce", primitive)
				}
			},
		},
		{
			name: "put on top of library",
			text: "When this creature attacks or blocks, put it on top of its owner's library at end of combat.",
			check: func(t *testing.T, primitive game.Primitive) {
				t.Helper()
				put, ok := primitive.(game.PutPermanentOnLibrary)
				if !ok || put.Object != game.SourceCardPermanentReference() || put.Bottom {
					t.Fatalf("primitive = %#v, want source-card put on top of library", primitive)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Fragile",
				Layout:     "normal",
				TypeLine:   "Creature — Elemental",
				OracleText: tt.text,
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
			}
			sequence := face.TriggeredAbilities[0].Content.Modes[0].Sequence
			if len(sequence) != 1 {
				t.Fatalf("sequence length = %d, want 1", len(sequence))
			}
			delayed, ok := sequence[0].Primitive.(game.CreateDelayedTrigger)
			if !ok {
				t.Fatalf("primitive = %#v, want delayed trigger", sequence[0].Primitive)
			}
			if delayed.Trigger.Timing != game.DelayedAtEndOfCombat {
				t.Fatalf("timing = %v, want DelayedAtEndOfCombat", delayed.Trigger.Timing)
			}
			tt.check(t, delayed.Trigger.Content.Modes[0].Sequence[0].Primitive)
		})
	}
}
