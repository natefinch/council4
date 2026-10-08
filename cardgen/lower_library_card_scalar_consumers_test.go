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
				reveal, ok := sequence[0].Primitive.(game.Reveal)
				if !ok {
					t.Fatalf("unexpected reveal primitive %T", sequence[0].Primitive)
				}
				outputs = reveal.PublishCharacteristics
			case game.PrimitiveLookAtLibraryTop:
				look, ok := sequence[0].Primitive.(game.LookAtLibraryTop)
				if !ok {
					t.Fatalf("unexpected look primitive %T", sequence[0].Primitive)
				}
				outputs = look.PublishCharacteristics
			default:
				t.Fatalf("unmodeled publisher %T", sequence[0].Primitive)
			}
			if outputs.Power != "sequence-effect-0-power" || outputs.Toughness != "sequence-effect-0-toughness" {
				t.Fatalf("scalar outputs=%#v", outputs)
			}
			for _, instruction := range sequence[1:] {
				var quantity game.Quantity
				switch instruction.Primitive.Kind() {
				case game.PrimitiveDraw:
					draw, ok := instruction.Primitive.(game.Draw)
					if !ok {
						t.Fatalf("unexpected draw primitive %T", instruction.Primitive)
					}
					quantity = draw.Amount
				case game.PrimitiveGainLife:
					gain, ok := instruction.Primitive.(game.GainLife)
					if !ok {
						t.Fatalf("unexpected gain-life primitive %T", instruction.Primitive)
					}
					quantity = gain.Amount
				case game.PrimitiveLoseLife:
					lose, ok := instruction.Primitive.(game.LoseLife)
					if !ok {
						t.Fatalf("unexpected lose-life primitive %T", instruction.Primitive)
					}
					quantity = lose.Amount
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
