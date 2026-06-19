package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerTriggerBodyResolutionCondition verifies that a triggered ability
// whose resolving body carries a state condition checked only on resolution
// ("Whenever X, EFFECT. If STATE, EFFECT2.") routes through the shared content
// lowering exactly as the same body lowers on a spell, rather than being
// rejected by the trigger-body preparation gate. The condition is not the
// trigger's intervening "if", so it stays in the body.
func TestLowerTriggerBodyResolutionCondition(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		OracleText: "When this creature enters, put a +1/+1 counter on this creature. If you control eight or more lands, draw a card.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("got %d triggered abilities, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Pattern.Event != game.EventPermanentEnteredBattlefield {
		t.Errorf("event = %v, want EventPermanentEnteredBattlefield", ta.Trigger.Pattern.Event)
	}
	if len(ta.Content.Modes) != 1 {
		t.Fatalf("modes = %#v, want 1", ta.Content.Modes)
	}
	seq := ta.Content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(seq))
	}
	if _, ok := seq[0].Primitive.(game.AddCounter); !ok {
		t.Errorf("instruction[0] = %#v, want AddCounter", seq[0].Primitive)
	}
	if _, ok := seq[1].Primitive.(game.Draw); !ok {
		t.Errorf("instruction[1] = %#v, want Draw", seq[1].Primitive)
	}
}
