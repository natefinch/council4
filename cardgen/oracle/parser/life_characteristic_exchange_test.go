package parser

import "testing"

func TestParseLifeCharacteristicExchange(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		cardName       string
		text           string
		kind           LifeCharacteristicExchangeKind
		targetOpponent bool
	}{
		{
			name:     "controller toughness",
			cardName: "Tree of Redemption",
			text:     "{T}: Exchange your life total with this creature's toughness.",
			kind:     LifeCharacteristicExchangeSourceToughness,
		},
		{
			name:           "opponent toughness",
			cardName:       "Tree of Perdition",
			text:           "{T}: Exchange target opponent's life total with this creature's toughness.",
			kind:           LifeCharacteristicExchangeSourceToughness,
			targetOpponent: true,
		},
		{
			name:     "controller power",
			cardName: "Evra, Halcyon Witness",
			text:     "{4}: Exchange your life total with Evra's power.",
			kind:     LifeCharacteristicExchangeSourcePower,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.text, Context{CardName: test.cardName})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			exchange := document.Abilities[0].LifeCharacteristicExchange
			if exchange == nil ||
				exchange.Kind != test.kind ||
				exchange.TargetOpponent != test.targetOpponent {
				t.Fatalf("exchange = %#v", exchange)
			}
		})
	}
}
