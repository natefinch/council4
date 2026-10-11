package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

func compiledSearchUnitContent(t *testing.T, text string) compiler.AbilityContent {
	t.Helper()
	doc, diagnostics := parser.Parse(text, parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	result, diagnostics := compiler.Compile(doc, compiler.Context{})
	if len(diagnostics) != 0 || len(result.Abilities) != 1 {
		t.Fatal(diagnostics)
	}
	return result.Abilities[0].Content
}

func TestOwnedSearchRequiresTypedOwnership(t *testing.T) {
	t.Parallel()
	foreign := compiledSearchUnitContent(t, ownedLandSearch)
	for _, tt := range []struct {
		name   string
		mutate func(*compiler.AbilityContent)
		valid  bool
	}{
		{"owned", func(*compiler.AbilityContent) {}, true},
		{"text and positions irrelevant", func(c *compiler.AbilityContent) {
			for i := range c.Effects {
				c.Effects[i].Text = "diagnostic only"
				c.Effects[i].Span = shared.Span{}
				c.Effects[i].ClauseSpan = shared.Span{}
				c.Effects[i].VerbSpan = shared.Span{}
			}
			c.References[0].Span = shared.Span{}
			c.Effects[1].References[0].Span = shared.Span{}
		}, true},
		{"missing group reference", func(c *compiler.AbilityContent) { c.Effects[1].References = nil }, false},
		{"foreign proof", func(c *compiler.AbilityContent) { c.Effects[1].References = foreign.Effects[1].References }, false},
		{"missing producer", func(c *compiler.AbilityContent) { c.Effects[1].References[0].ProducerClauseID = 0 }, false},
		{"forward producer", func(c *compiler.AbilityContent) { c.Effects[1].References[0].PriorInstruction = 2 }, false},
		{"target binding", func(c *compiler.AbilityContent) { c.Effects[1].References[0].Binding = compiler.ReferenceBindingTarget }, false},
		{"duplicate reference", func(c *compiler.AbilityContent) {
			c.Effects[1].References = append(c.Effects[1].References, c.Effects[1].References[0])
		}, false},
		{"missing clause", func(c *compiler.AbilityContent) { c.Effects[1].ClauseID = 0 }, false},
		{"duplicate clause", func(c *compiler.AbilityContent) { c.Effects[1].ClauseID = c.Effects[0].ClauseID }, false},
		{"unowned source order", func(c *compiler.AbilityContent) { c.Effects[1].VerbOrder = shared.SourceOrder{} }, false},
		{"changed canonical reference order", func(c *compiler.AbilityContent) { c.Effects[1].References[0].Order.Start++ }, false},
		{"internal condition", func(c *compiler.AbilityContent) {
			c.Conditions = []compiler.CompiledCondition{{
				Kind: compiler.ConditionIf, Predicate: compiler.ConditionPredicateControllerHandEmpty,
				Ownership: parser.ConditionOwnership{Scope: parser.ConditionScopeClause, ClauseIDs: []int{c.Effects[1].ClauseID}},
			}}
		}, false},
		{"zero quantity", func(c *compiler.AbilityContent) { c.Effects[0].Amount.Value = 0 }, false},
		{"internal optional", func(c *compiler.AbilityContent) { c.Effects[1].Optional = true }, false},
		{"internal branch", func(c *compiler.AbilityContent) { c.Effects[1].Connection = parser.EffectConnectionOtherwise }, false},
		{"shuffle zone movement", func(c *compiler.AbilityContent) {
			c.Effects[2].FromZone = zone.Graveyard
			c.Effects[2].ToZone = zone.Library
		}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			content := compiledSearchUnitContent(t, ownedLandSearch)
			content.Effects[1].References = slices.Clone(content.Effects[1].References)
			tt.mutate(&content)
			units, reason := planSearchSequenceUnits(content)
			if (reason == "") != tt.valid || tt.valid && len(units) != 1 {
				t.Fatalf("units=%d reason=%q, want valid=%v", len(units), reason, tt.valid)
			}
		})
	}
}

func TestOwnedSearchRefusals(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Draw a card. Search your library for zero basic land cards, put them onto the battlefield, then shuffle. Scry 1.",
		"Draw a card. Search your library for up to X basic land cards, where X is the number of creatures an unknown player controls, put them onto the battlefield, then shuffle. Scry 1.",
		"Draw a card. Search target opponent's library for a basic land card, put it onto the battlefield, then shuffle. Scry 1.",
		"Draw a card. Search your library for a basic land card, put it onto the battlefield tapped if you control a creature, then shuffle. Scry 1.",
		"Draw a card. Search your library for a basic land card, put it onto the battlefield tapped and angry, then shuffle. Scry 1.",
		"Draw a card. Search your library for a basic land card, put it onto the battlefield tapped, win the game, then shuffle. Scry 1.",
		"Draw a card. Search your library for a basic land card, put it onto the battlefield tapped, then shuffle your graveyard into your library, then scry 1.",
		ownedLandSearch + " Untap that creature.",
		ownedLandSearch + " It gains haste and banding or flying until end of turn.",
		"Untap that land. " + ownedLandSearch,
		"Search your library for two basic land cards, put them onto the battlefield tapped, then shuffle. Untap that land.",
		"Search your library for a basic land card, put it into your hand, then shuffle. Untap that land.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, &ScryfallCard{Name: "Search Refusal", Layout: "normal", TypeLine: "Sorcery", OracleText: text})
		})
	}

}
