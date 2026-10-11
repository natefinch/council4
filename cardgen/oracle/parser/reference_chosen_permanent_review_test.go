package parser

import "testing"

func TestChosenPermanentNameUnmodeledChoiceGrammarRefuses(t *testing.T) {
	for _, choice := range []string{
		"Choose a land at random from among lands you control. ",
		"Choose a land secretly from among lands you control. ",
	} {
		t.Run(choice, func(t *testing.T) {
			document, diagnostics := Parse(choice+"You gain 2 life. "+chosenNameSearchText, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			for _, reference := range document.Abilities[0].SemanticReferences {
				if reference.Text == "the chosen land" && reference.ChoiceClauseID != 0 {
					t.Fatal("unmodeled choice grammar acquired canonical original-choice authority")
				}
			}
		})
	}
}

func TestChosenPermanentNameClosedChoiceGrammarPreservesQualifiers(t *testing.T) {
	for _, choice := range []string{
		"Choose a land you control. ",
		"Choose a nonbasic land you control. ",
		"Choose a tapped land you control. ",
		"Choose an artifact you control. ",
		"Choose a creature you control. ",
	} {
		t.Run(choice, func(t *testing.T) {
			document, diagnostics := Parse(choice, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			effects := conditionEffects(document.Abilities[0].Sentences)
			if len(effects) != 1 || effects[0].Kind != EffectChoosePermanent || !effects[0].Exact {
				t.Fatalf("closed typed choice grammar lost supported selection qualifier: %+v", effects[0].Selection)
			}
		})
	}
}
