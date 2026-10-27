// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestDepartureCaptureWiring(t *testing.T) {
	for _, tc := range []struct {
		def   *game.CardDef
		power bool
	}{
		{RTDepartureCapture(), true},
		{RTDepartureToughness(), false},
	} {
		sequence := tc.def.TriggeredAbilities[0].Content.Modes[0].Sequence
		if len(sequence) != 2 {
			t.Fatalf("sequence = %+v, want move and consumer", sequence)
		}
		move := sequence[0].Primitive.(game.MoveCard)
		key, other := move.PublishDepartureCharacteristics.Power, move.PublishDepartureCharacteristics.Toughness
		if !tc.power {
			key, other = other, key
		}
		if key == "" || other != "" || !sequence[0].LocalProducts.HasResult(key) {
			t.Fatalf("generated capture/local declaration lost: %+v", sequence[0])
		}
		gain := sequence[1].Primitive.(game.GainLife)
		amount := gain.Amount.DynamicAmount()
		if !amount.Exists || amount.Val.Kind != game.DynamicAmountPreviousEffectResult ||
			amount.Val.ResultKey != key || !sequence[1].ResultGate.Exists ||
			sequence[1].ResultGate.Val.Key != key || !sequence[1].ResultGate.Val.AmountAvailable {
			t.Fatalf("generated consumer lost exact scalar or availability: %+v", sequence[1])
		}
	}
}
