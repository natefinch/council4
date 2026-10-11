package parser

import "testing"

func TestActualCardQuantityGrammarAndOwnership(t *testing.T) {
	for _, tc := range []struct {
		text       string
		producer   int
		outcome    EffectKind
		multiplier int
	}{
		{"Exile all graveyards. For each creature card exiled this way, create a 2/2 black Zombie creature token.", 1, EffectExile, 1},
		{"Exile three cards from your graveyard. For each creature card exiled this way, create a 2/2 black Zombie creature token.", 1, EffectExile, 1},
		{"Exile all graveyards. Draw a card. For each artifact or land card exiled this way, you create a 1/1 green Insect creature token.", 1, EffectExile, 1},
		{"Mill three cards. For each creature card milled this way, draw two cards.", 1, EffectMill, 2},
		{"Discard two cards. For each nonland card discarded this way, you gain 2 life.", 1, EffectDiscard, 2},
		{"Mill two cards. Draw a card. Mill three cards. For each creature card milled this way, draw a card.", 3, EffectMill, 1},
		{"For each creature card milled this way, draw a card. Mill three cards.", 0, EffectMill, 1},
	} {
		t.Run(tc.text, func(t *testing.T) {
			doc, diagnostics := Parse(tc.text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics: %#v", diagnostics)
			}
			var quantities []EffectSyntax
			for _, sentence := range doc.Abilities[0].Sentences {
				for _, effect := range sentence.Effects {
					if effect.Kind == EffectExile && tc.producer > 0 && !effect.Exact {
						t.Errorf("producer is not exact: %q", exactEffectClauseText(&effect))
					}
					if effect.Amount.DynamicKind == EffectDynamicAmountActualCardResultCount {
						quantities = append(quantities, effect)
					}
				}
			}
			if len(quantities) != 1 {
				t.Fatalf("quantities=%#v; sentences=%#v", quantities, doc.Abilities[0].Sentences)
			}
			effect := quantities[0]
			amount := effect.Amount
			if effect.Exact != (tc.producer > 0) || amount.ProducerClauseID != tc.producer || amount.ResultOutcome != tc.outcome ||
				amount.Multiplier != tc.multiplier || amount.Selection == nil || amount.Known || amount.VariableX {
				t.Fatalf("effect exact=%t amount=%#v; want producer=%d outcome=%s multiplier=%d",
					effect.Exact, amount, tc.producer, tc.outcome, tc.multiplier)
			}
		})
	}
}
