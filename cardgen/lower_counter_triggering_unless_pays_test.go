package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerCounterTriggeringSpellOrAbilityUnlessPays(t *testing.T) {
	t.Parallel()
	const oracleText = "Whenever this creature becomes the target of a spell or ability an opponent controls, counter that spell or ability unless its controller pays {2}."
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Frost Titan",
		Layout:     "normal",
		TypeLine:   "Creature — Giant",
		OracleText: oracleText,
	})
	sequence := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence length = %d, want pay then counter", len(sequence))
	}
	pay, ok := sequence[0].Primitive.(game.Pay)
	if !ok || !pay.Payment.Payer.Exists {
		t.Fatalf("instruction 0 = %#v, want payment by stack-object controller", sequence[0])
	}
	object, ok := pay.Payment.Payer.Val.Object()
	if !ok || object != game.EventStackObjectReference() {
		t.Fatalf("payer = %#v, want event stack object's controller", pay.Payment.Payer)
	}
	counter, ok := sequence[1].Primitive.(game.CounterObject)
	if !ok || counter.Object != game.EventStackObjectReference() {
		t.Fatalf("instruction 1 = %#v, want event-stack counter", sequence[1])
	}
}
