package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLibraryCardCharacteristicConsumersUseSeparateAvailableScalars(t *testing.T) {
	for _, text := range []string{
		"Reveal the top card of your library. If it's a creature card, you draw cards equal to its power and you gain life equal to its toughness.",
		"Look at the top card of your library. If it's a creature card, you gain life equal to that card's toughness and you lose life equal to its power.",
	} {
		t.Run(text, func(t *testing.T) {
			card := &ScryfallCard{Name: "Observed Characteristics", Layout: "normal", TypeLine: "Sorcery", OracleText: text}
			face := lowerSingleFace(t, card)
			sequence := face.SpellAbility.Val.Modes[0].Sequence
			if len(sequence) != 3 {
				t.Fatalf("instructions=%d, want observation and two consumers", len(sequence))
			}
			outputs := game.LibraryCardCharacteristics{}
			switch sequence[0].Primitive.Kind() {
			case game.PrimitiveReveal:
				outputs = sequence[0].Primitive.(game.Reveal).PublishCharacteristics
			case game.PrimitiveLookAtLibraryTop:
				outputs = sequence[0].Primitive.(game.LookAtLibraryTop).PublishCharacteristics
			}
			if outputs.Power != "sequence-effect-0-power" || outputs.Toughness != "sequence-effect-0-toughness" {
				t.Fatalf("scalar outputs=%#v", outputs)
			}
			for _, instruction := range sequence[1:] {
				var quantity game.Quantity
				switch instruction.Primitive.Kind() {
				case game.PrimitiveDraw:
					quantity = instruction.Primitive.(game.Draw).Amount
				case game.PrimitiveGainLife:
					quantity = instruction.Primitive.(game.GainLife).Amount
				case game.PrimitiveLoseLife:
					quantity = instruction.Primitive.(game.LoseLife).Amount
				default:
					t.Fatalf("unmodeled consumer %T", instruction.Primitive)
				}
				dynamic := quantity.DynamicAmount()
				if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountPreviousEffectResult ||
					!instruction.ResultGate.Exists || !instruction.ResultGate.Val.AmountAvailable ||
					instruction.ResultGate.Val.Key != dynamic.Val.ResultKey ||
					dynamic.Val.ResultKey != outputs.Power && dynamic.Val.ResultKey != outputs.Toughness {
					t.Fatalf("consumer=%#v amount=%#v", instruction, dynamic)
				}
			}
		})
	}
}
