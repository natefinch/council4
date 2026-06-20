package parser

import "testing"

func TestParseImpulseExileGeneralizedForms(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		text     string
		amount   int
		duration EffectDurationKind
	}{
		{
			name:     "single card this turn",
			text:     "Exile the top card of your library. You may play that card this turn.",
			amount:   1,
			duration: EffectDurationThisTurn,
		},
		{
			name:     "single card it until end of turn",
			text:     "Exile the top card of your library. You may play it until end of turn.",
			amount:   1,
			duration: EffectDurationUntilEndOfTurn,
		},
		{
			name:     "single card leading until end of turn",
			text:     "Exile the top card of your library. Until end of turn, you may play that card.",
			amount:   1,
			duration: EffectDurationUntilEndOfTurn,
		},
		{
			name:     "two cards until end of your next turn",
			text:     "Exile the top two cards of your library. You may play those cards until the end of your next turn.",
			amount:   2,
			duration: EffectDurationUntilEndOfYourNextTurn,
		},
		{
			name:     "three cards leading until end of your next turn",
			text:     "Exile the top three cards of your library. Until the end of your next turn, you may play those cards.",
			amount:   3,
			duration: EffectDurationUntilEndOfYourNextTurn,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(tc.text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 || len(document.Abilities) != 1 {
				t.Fatalf("diagnostics = %#v, abilities = %#v", diagnostics, document.Abilities)
			}
			effect := document.Abilities[0].Sentences[0].Effects[0]
			if effect.Kind != EffectImpulseExile || !effect.Exact ||
				!effect.Amount.Known || effect.Amount.Value != tc.amount ||
				effect.Duration != tc.duration {
				t.Fatalf("impulse effect = %#v", effect)
			}
		})
	}
}

func TestParseImpulseExileFailsClosed(t *testing.T) {
	t.Parallel()
	variants := []string{
		"Exile the top card of your library. You may play it.",
		"Exile the top card of your library. You may play it until your next turn.",
		"Exile the top three cards of your library. You may play that card this turn.",
		"Exile the top card of your library. You may play them this turn.",
	}
	for _, source := range variants {
		document, _ := Parse(source, Context{InstantOrSorcery: true})
		if len(document.Abilities) == 1 &&
			len(document.Abilities[0].Sentences) > 0 &&
			len(document.Abilities[0].Sentences[0].Effects) > 0 &&
			document.Abilities[0].Sentences[0].Effects[0].Kind == EffectImpulseExile {
			t.Fatalf("variant was recognized as impulse exile:\n%s", source)
		}
	}
}
