package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestChosenNamePublicIndependentClauses(t *testing.T) {
	const search = "Search your library for up to two basic land cards which have the same name as the chosen land, put them onto the battlefield tapped, then shuffle."
	for _, tc := range []struct {
		text     string
		searches int
	}{
		{"Choose a land you control. " + search, 1},
		{"Choose a land you control. You gain 2 life. " + search, 1},
		{"You gain 2 life. Choose a land you control. " + search, 1},
		{"Choose a land you control. " + search + " You gain 2 life.", 1},
		{"Choose a land you control. You gain 2 life. " + search + " You gain 3 life. " + search, 2},
	} {
		t.Run(tc.text, func(t *testing.T) {
			card := &ScryfallCard{Name: "Chosen name composition", Layout: "normal", TypeLine: "Sorcery", OracleText: tc.text}
			defs, diagnostics, err := CompileCardDefs(card)
			if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
				t.Fatalf("public compile: defs=%d diagnostics=%+v err=%v", len(defs), diagnostics, err)
			}
			choices, searches := 0, 0
			for _, instruction := range defs[0].SpellAbility.Val.Modes[0].Sequence {
				if choose, ok := instruction.Primitive.(game.Choose); ok {
					choices++
					if choose.Choice.Kind != game.ResolutionChoicePermanent ||
						choose.PublishChoice != game.ResolutionChosenPermanentChoiceKey {
						t.Fatalf("changed original choice: %+v", choose)
					}
				}
				if search, ok := instruction.Primitive.(game.Search); ok {
					searches++
					if search.Spec.NameFromChoice != game.ResolutionChosenPermanentChoiceKey || search.Amount.Value() != 2 {
						t.Fatalf("changed search operand: %+v", search)
					}
				}
			}
			if choices != 1 || searches != tc.searches {
				t.Fatalf("choices/searches=%d/%d", choices, searches)
			}
		})
	}
}

func TestChosenNamePublicMissingAmbiguousAndSkippedChoiceRefuse(t *testing.T) {
	const search = "Search your library for up to two basic land cards which have the same name as the chosen land, put them onto the battlefield tapped, then shuffle."
	for _, prefix := range []string{
		"",
		"Choose a creature you control. ",
		"Choose a land you control. Choose a land you control. ",
		"Choose a land you control. Bolster 1. ",
		"If you control a creature, choose a land you control. ",
		"You may choose a land you control. ",
	} {
		t.Run(prefix, func(t *testing.T) {
			card := &ScryfallCard{Name: "Chosen name refusal", Layout: "normal", TypeLine: "Sorcery", OracleText: prefix + search}
			assertCardUnsupported(t, card)
			defs, diagnostics, err := CompileCardDefs(card)
			if err != nil || len(diagnostics) == 0 || len(defs) != 0 {
				t.Fatal("refused chosen-name authority produced a partial CardDef")
			}
		})
	}
}
