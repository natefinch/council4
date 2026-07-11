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
		{
			name:     "single card it until your next end step",
			text:     "Exile the top card of your library. Until your next end step, you may play it.",
			amount:   1,
			duration: EffectDurationUntilYourNextEndStep,
		},
		{
			name:     "single card that card until your next end step",
			text:     "Exile the top card of your library. You may play that card until your next end step.",
			amount:   1,
			duration: EffectDurationUntilYourNextEndStep,
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

// TestParseImpulseExileIgnoresTrailingReminder confirms that trailing reminder
// text ("(If you cast a spell this way…)") does not block impulse recognition
// (Act on Impulse, Dark-Dweller Oracle).
func TestParseImpulseExileIgnoresTrailingReminder(t *testing.T) {
	t.Parallel()
	source := "Exile the top three cards of your library. Until end of turn, you may play those cards. (If you cast a spell this way, you still pay its costs. You can play a land this way only if you have an available land play remaining.)"
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("diagnostics = %#v, abilities = %#v", diagnostics, document.Abilities)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if effect.Kind != EffectImpulseExile || !effect.Exact ||
		!effect.Amount.Known || effect.Amount.Value != 3 ||
		effect.Duration != EffectDurationUntilEndOfTurn {
		t.Fatalf("impulse effect = %#v", effect)
	}
}

// TestParseImpulseExileVariableX confirms that "Exile the top X cards of your
// library." carries the spell's {X} as the impulse amount (Commune with Lava,
// Hugs, Grisly Guardian).
func TestParseImpulseExileVariableX(t *testing.T) {
	t.Parallel()
	source := "Exile the top X cards of your library. Until the end of your next turn, you may play those cards."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 || len(document.Abilities) != 1 {
		t.Fatalf("diagnostics = %#v, abilities = %#v", diagnostics, document.Abilities)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if effect.Kind != EffectImpulseExile || !effect.Exact ||
		effect.Amount.Known || !effect.Amount.VariableX ||
		effect.Duration != EffectDurationUntilEndOfYourNextTurn {
		t.Fatalf("impulse effect = %#v", effect)
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

// TestFoldTrailingImpulseExileFailsClosedOnCondition confirms the trailing
// impulse-exile fold does not fire when the exile sentence carries a condition,
// whether leading or trailing. Folding collapses the body to the pure impulse
// shape whose condition clauses are then stripped, so a gated exile must stay
// unfolded (and unsupported) rather than generate an ungated impulse.
func TestFoldTrailingImpulseExileFailsClosedOnCondition(t *testing.T) {
	t.Parallel()
	variants := []string{
		"Draw a card. Exile the top card of your library if you control an artifact. You may play it this turn.",
		"Draw a card. If you control an artifact, exile the top card of your library. You may play it this turn.",
		"Draw a card. Exile the top card of your library unless you control an artifact. You may play it this turn.",
	}
	for _, source := range variants {
		document, _ := Parse(source, Context{InstantOrSorcery: true})
		for _, ability := range document.Abilities {
			for _, sentence := range ability.Sentences {
				for _, effect := range sentence.Effects {
					if effect.Kind == EffectImpulseExile {
						t.Fatalf("gated exile was folded into an ungated impulse:\n%s", source)
					}
				}
			}
		}
	}
}
