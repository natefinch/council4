package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

func TestLowerClockworkEndOfCombatCounterRemoval(t *testing.T) {
	t.Parallel()
	const oracleText = "Whenever this creature attacks or blocks, remove a +1/+1 counter from it at end of combat."
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Clockwork Beetle",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Insect",
		OracleText: oracleText,
	})
	sequence := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	delayed, ok := sequence[0].Primitive.(game.CreateDelayedTrigger)
	if !ok ||
		delayed.Trigger.Timing != game.DelayedAtEndOfCombat ||
		!delayed.Trigger.CapturedObject.Exists ||
		delayed.Trigger.CapturedObject.Val != game.EventPermanentReference() {
		t.Fatalf("instruction = %#v, want captured end-of-combat trigger", sequence[0])
	}
	remove, ok := delayed.Trigger.Content.Modes[0].Sequence[0].Primitive.(game.RemoveCounter)
	if !ok ||
		remove.Object != game.CapturedObjectReference() ||
		remove.CounterKind != counter.PlusOnePlusOne ||
		remove.Amount.Value() != 1 {
		t.Fatalf("delayed primitive = %#v, want one captured +1/+1 counter removal", delayed.Trigger.Content)
	}
}
