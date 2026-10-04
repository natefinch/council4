package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerResultSelectionIsTextBlind(t *testing.T) {
	document, diagnostics := parser.Parse("Discard a card. If a land card is discarded this way, draw a card.", parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	compilation, diagnostics := compiler.Compile(document, compiler.Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	content := compilation.Abilities[0].Content
	content.Conditions[0].Text = "if a creature is destroyed this way"
	for i := range content.Effects {
		content.Effects[i].Text = "opaque"
	}
	plan, ok := planOptionalFlow(content)
	if !ok || !plan.resultSelection.Exists || !plan.resultCardNoun ||
		!reflect.DeepEqual(plan.resultSelection.Val, game.Selection{RequiredTypes: []types.Card{types.Land}}) {
		t.Fatalf("plan = %#v, %t", plan, ok)
	}
	content.Conditions[0].ThisWaySelection = nil
	if _, ok := planOptionalFlow(content); ok {
		t.Fatal("missing typed payload fell back to retained text or generic success")
	}
}

func TestLowerResultSelectionCardPaths(t *testing.T) {
	windgrace := &ScryfallCard{
		Name: "Lord Windgrace", Layout: "normal", TypeLine: "Legendary Planeswalker — Windgrace",
		ManaCost: "{3}{B}{R}{G}", Loyalty: new("4"),
		OracleText: "+2: Discard a card, then draw a card. If a land card is discarded this way, draw an additional card.",
	}
	assertCardPaths(t, windgrace,
		"LoyaltyAbilities[0].Content.Modes[0].Sequence[0].PublishResult",
		"LoyaltyAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.Discard).Amount.fixed = 1",
		"LoyaltyAbilities[0].Content.Modes[0].Sequence[2].ResultGate.Val.Succeeded = game.TriTrue",
		"LoyaltyAbilities[0].Content.Modes[0].Sequence[2].ResultGate.Val.ObjectSelection.Val.RequiredTypes[0] = types.Land",
		"LoyaltyAbilities[0].Content.Modes[0].Sequence[2].ResultGate.Val.CardOnly = true",
	)
	assertCardPathsAbsent(t, windgrace,
		"LoyaltyAbilities[0].Content.Modes[0].Sequence[1].ResultGate.Val.Key",
	)
	ruse := &ScryfallCard{
		Name: "Siren's Ruse", Layout: "normal", TypeLine: "Instant",
		OracleText: "Exile target creature you control, then return that card to the battlefield under its owner's control. If a Pirate was exiled this way, draw a card.",
	}
	assertCardPaths(t, ruse,
		"Sequence[0].Primitive.(game.MovePermanent).PublishLinked = \"blink-1\"",
		"Sequence[2].ResultGate.Val.ObjectSelection.Val.SubtypesAny[0] = types.Pirate",
	)
	literal := &ScryfallCard{
		Name: "Literal Result", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "You may discard a card. If you do, draw a card.",
	}
	assertCardPathsAbsent(t, literal, "ObjectSelection.Exists = true")
}

func TestLowerResultSelectionCardDomainFailsClosedOnPermanentDeparture(t *testing.T) {
	assertCardUnsupported(t, &ScryfallCard{
		Name: "Card Departure Probe", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "Exile target creature. If a creature card was exiled this way, draw a card.",
	}, "does not publish post-move card objects")
}

func TestLowerResultSelectionOtherwiseCardPaths(t *testing.T) {
	assertCardPaths(t, &ScryfallCard{
		Name: "Otherwise Probe", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "You may discard a card. If a land card is discarded this way, draw two cards. Otherwise, draw a card.",
	},
		"Sequence[1].ResultGate.Val.ObjectSelection.Val.RequiredTypes[0] = types.Land",
		"Sequence[2].ResultGate.Val.ObjectSelection.Val.RequiredTypes[0] = types.Land",
		"Sequence[2].ResultGate.Val.CardOnly = true",
		"Sequence[2].ResultGate.Val.Succeeded = game.TriTrue",
		"Sequence[2].ResultGate.Val.Negate = true",
	)
}

func TestLowerResultSelectionExplicitIfDontIsUnfiltered(t *testing.T) {
	card := &ScryfallCard{
		Name: "Explicit Decline Probe", Layout: "normal", TypeLine: "Sorcery",
		OracleText: "You may discard a card. If a land card is discarded this way, draw two cards. If you don't, draw a card.",
	}
	assertCardPaths(t, card, "Sequence[2].ResultGate.Val.Succeeded = game.TriFalse")
	assertCardPathsAbsent(t, card,
		"Sequence[2].ResultGate.Val.ObjectSelection.Exists = true",
		"Sequence[2].ResultGate.Val.CardOnly = true",
		"Sequence[2].ResultGate.Val.Negate = true",
	)
}
