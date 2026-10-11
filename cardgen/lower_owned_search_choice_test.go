package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
)

const ownedNameSearch = "Search your library for up to two basic land cards which have the same name as the chosen land, put them onto the battlefield tapped, then shuffle."
const ownedChoiceSearch = "Choose a land you control. " + ownedNameSearch

func TestOwnedSearchNameChoice(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		ownedChoiceSearch,
		"You gain 2 life. " + ownedChoiceSearch + " Scry 1.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			assertCardPaths(t, &ScryfallCard{Name: "Name Choice", Layout: "normal", TypeLine: "Sorcery", OracleText: text},
				"Primitive.(game.Choose).Choice.Kind = game.ResolutionChoicePermanent",
				"Primitive.(game.Choose).PublishChoice = game.ResolutionChosenPermanentChoiceKey",
				"Primitive.(game.Search).Spec.NameFromChoice = game.ResolutionChosenPermanentChoiceKey",
				"Primitive.(game.Search).Amount.fixed = 2",
			)
		})
	}
}

func TestOwnedSearchNameChoicePrivateProof(t *testing.T) {
	t.Parallel()
	foreign := compiledSearchUnitContent(t, ownedChoiceSearch)
	for _, tt := range []struct {
		name   string
		mutate func(*compiler.AbilityContent)
		valid  bool
	}{
		{"owned", func(*compiler.AbilityContent) {}, true},
		{"foreign", func(c *compiler.AbilityContent) { c.Effects[1].References = foreign.Effects[1].References }, false},
		{"missing", func(c *compiler.AbilityContent) { c.Effects[1].References[0].ProducerClauseID = 0 }, false},
		{"forward", func(c *compiler.AbilityContent) { c.Effects[1].References[0].PriorInstruction = 2 }, false},
		{"non-choice", func(c *compiler.AbilityContent) { c.Effects[0].Kind = compiler.EffectDraw }, false},
		{"duplicate ClauseID", func(c *compiler.AbilityContent) { c.Effects[0].ClauseID = c.Effects[1].ClauseID }, false},
		{"duplicate reference", func(c *compiler.AbilityContent) {
			c.Effects[1].References = append(c.Effects[1].References, c.Effects[1].References[0])
		}, false},
		{"canonical order", func(c *compiler.AbilityContent) { c.Effects[1].References[0].Order.Start++ }, false},
		{"noun proof", func(c *compiler.AbilityContent) { c.Effects[1].References[0].SubjectNoun = "ObjectNounCreature" }, false},
		{"recognition flag", func(c *compiler.AbilityContent) { c.Effects[1].SearchSameNameAsChosenObject = false }, false},
		{"binding flag", func(c *compiler.AbilityContent) { c.Effects[1].SearchNameFromChosenPermanent = false }, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			content := compiledSearchUnitContent(t, ownedChoiceSearch)
			content.Effects[1].References = slices.Clone(content.Effects[1].References)
			tt.mutate(&content)
			_, reason := planSearchSequenceUnits(content)
			if (reason == "") != tt.valid {
				t.Fatalf("reason=%q, want valid=%v", reason, tt.valid)
			}
		})
	}
	content := compiledSearchUnitContent(t, ownedChoiceSearch)
	reference := content.Effects[1].References[0]
	for _, key := range []game.ChoiceKey{"", "foreign-choice"} {
		if searchNameChoiceOperand(content, 1, game.SearchSpec{NameFromChoice: key}, reference) {
			t.Fatal("an unrelated choice key established name-correlation authority")
		}
	}
}

func TestOwnedSearchNameChoiceRefusals(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Draw a card. " + ownedNameSearch,
		ownedNameSearch + " Choose a land you control.",
		"Choose two lands you control. " + ownedNameSearch,
		"Choose a land an opponent controls. " + ownedNameSearch,
		"Choose a land you control. You gain 2 life. " + ownedNameSearch,
		"If you control a creature, choose a land you control. " + ownedNameSearch,
		"Choose a land you control. Choose a land you control. " + ownedNameSearch,
		"Choose a land you control. Search your library for a card with the same color as the chosen land, put it into your hand, then shuffle.",
	} {
		assertCardUnsupported(t, &ScryfallCard{Name: "Choice Refusal", Layout: "normal", TypeLine: "Sorcery", OracleText: text})
	}
}
