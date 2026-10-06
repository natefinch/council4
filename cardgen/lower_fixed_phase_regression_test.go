package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestFixedPhaseExistingCompositionsRemainExecutable(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, typ, text string }{
		{"source blink", "Creature", "{1}: Exile this creature. Return it to the battlefield under its owner's control at the beginning of the next end step."},
		{"actual plural batch", "Sorcery", "Create three 1/1 red Human creature tokens with haste. Sacrifice those tokens at the beginning of the next end step."},
		{"keyword entry counter", "Sorcery", "Exile up to one target creature you control without flying. Return it to the battlefield under its owner's control with a flying counter on it at the beginning of the next end step."},
		{"source already in graveyard", "Creature", "Whenever a creature enters, if it entered from your graveyard or you cast it from your graveyard, return this card from your graveyard to the battlefield tapped at the beginning of the next end step."},
		{"established optional fallback", "Creature", "{2}, {T}: Exile another target creature you control. You may return that card to the battlefield under its owner's control. If you don't, at the beginning of the next end step, return that card to the battlefield under its owner's control with a vigilance counter and a lifelink counter on it."},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			defs, diagnostics, err := CompileCardDefs(&ScryfallCard{
				Name: "Existing Composition", Layout: "normal", TypeLine: test.typ, OracleText: test.text,
				Power: new("2"), Toughness: new("2"),
			})
			if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
				t.Fatalf("compile: %v; diagnostics=%+v; definitions=%d", err, diagnostics, len(defs))
			}
			if problems := game.ValidateCardDef(defs[0]); len(problems) != 0 {
				t.Fatalf("nonexecutable capture: %+v", problems)
			}
		})
	}
}
