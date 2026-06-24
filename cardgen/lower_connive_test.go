package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerConniveTrigger proves Ledger Shredder's "Whenever a player casts
// their second spell each turn, this creature connives." lowers to a single
// triggered ability whose body is a game.Connive primitive. The connive keyword
// action (CR 702.154: draw a card, then discard a card; if the discarded card
// was a nonland, put a +1/+1 counter on this creature) is modeled as one
// runtime primitive scoped to the source permanent and its controller. The
// printed reminder text is subsumed by the runtime handler and adds nothing to
// the lowered sequence.
func TestLowerConniveTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Ledger Shredder",
		Layout:   "normal",
		TypeLine: "Creature — Bird Advisor",
		OracleText: "Flying\nWhenever a player casts their second spell each turn, " +
			"this creature connives. (Draw a card, then discard a card. If you " +
			"discarded a nonland card, put a +1/+1 counter on this creature.)",
		Power:     new("1"),
		Toughness: new("3"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ab := face.TriggeredAbilities[0]
	if ab.Trigger.Pattern.Event != game.EventSpellCast {
		t.Fatalf("trigger event = %v, want EventSpellCast", ab.Trigger.Pattern.Event)
	}
	if ab.Trigger.Pattern.PlayerEventOrdinalThisTurn != 2 {
		t.Fatalf("trigger ordinal = %d, want 2", ab.Trigger.Pattern.PlayerEventOrdinalThisTurn)
	}
	mode := ab.Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("instruction count = %d, want 1", len(mode.Sequence))
	}
	connive, ok := mode.Sequence[0].Primitive.(game.Connive)
	if !ok {
		t.Fatalf("primitive = %#v, want game.Connive", mode.Sequence[0].Primitive)
	}
	if connive.Object != game.SourcePermanentReference() {
		t.Fatalf("connive object = %#v, want source permanent", connive.Object)
	}
	if connive.Player != game.ControllerReference() {
		t.Fatalf("connive player = %#v, want controller", connive.Player)
	}
	if connive.Amount != game.Fixed(1) {
		t.Fatalf("connive amount = %#v, want Fixed(1)", connive.Amount)
	}
}
