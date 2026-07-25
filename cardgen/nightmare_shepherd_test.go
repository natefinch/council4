package cardgen

import (
	"go/parser"
	"go/token"
	"testing"
)

const nightmareShepherdOracleText = "Flying\nWhenever another nontoken creature you control dies, you may exile it. If you do, create a token that's a copy of that creature, except it's 1/1 and it's a Nightmare in addition to its other types."

func nightmareShepherdCard() *ScryfallCard {
	return &ScryfallCard{
		Name:       "Nightmare Shepherd",
		Layout:     "normal",
		ManaCost:   "{2}{B}{B}",
		TypeLine:   "Enchantment Creature — Demon",
		OracleText: nightmareShepherdOracleText,
		Power:      new("4"),
		Toughness:  new("4"),
	}
}

// TestNightmareShepherdComposesEventExileAndLKICopy proves the card's three
// capabilities compose into one CardDef: a died trigger restricted to another
// nontoken creature you control, an optional exile of the card the event names
// (reached in the graveyard, because the permanent has already left the
// battlefield), and an "if you do" gate on a token that copies the dead
// creature except for its printed 1/1 body and added Nightmare type.
//
// The copy specification lives in game.CreateToken's unexported TokenSource, so
// these assertions are only meaningful because the dumper walks unexported
// state; see TestCardDefPathsWalkAllReachableState.
func TestNightmareShepherdComposesEventExileAndLKICopy(t *testing.T) {
	t.Parallel()
	card := nightmareShepherdCard()
	assertCardPaths(t, card,
		"StaticAbilities[0].KeywordAbilities[0].(game.SimpleKeyword).Kind = game.Flying",
		// "Whenever another nontoken creature you control dies".
		"Trigger.Pattern.Event = game.EventPermanentDied",
		"Trigger.Pattern.Controller = game.TriggerControllerYou",
		"Trigger.Pattern.ExcludeSelf = true",
		"Trigger.Pattern.SubjectSelection.RequiredTypes[0] = types.Creature",
		"Trigger.Pattern.SubjectSelection.NonToken = true",
		// "you may exile it": the card the event names, taken from the
		// graveyard it has already reached, publishing a link the copy reads.
		"Sequence[0].Primitive.(game.MoveCard).Card.Kind = game.CardReferenceEvent",
		"Sequence[0].Primitive.(game.MoveCard).FromZone = zone.Graveyard",
		"Sequence[0].Primitive.(game.MoveCard).Destination = zone.Exile",
		`Sequence[0].Primitive.(game.MoveCard).PublishLinked = "event-card-exile-copy"`,
		"Sequence[0].Primitive.(game.MoveCard).ReplacePublishedLinked = true",
		"Sequence[0].Primitive.(game.MoveCard).IncludeEventPermanentComponents = true",
		"Sequence[0].Optional = true",
		// "If you do": the token is created only when the exile succeeded.
		"Sequence[1].ResultGate.Val.Succeeded = game.TriTrue",
		// "a token that's a copy of that creature, except it's 1/1 and it's a
		// Nightmare in addition to its other types".
		"Sequence[1].Primitive.(game.CreateToken).Source.copy.Source = game.TokenCopySourceObject",
		"Sequence[1].Primitive.(game.CreateToken).Source.copy.Object.kind = game.ObjectReferenceLinkedObject",
		`Sequence[1].Primitive.(game.CreateToken).Source.copy.Object.linkID = "event-card-exile-copy"`,
		"Sequence[1].Primitive.(game.CreateToken).Source.copy.SetPower.Val.Value = 1",
		"Sequence[1].Primitive.(game.CreateToken).Source.copy.SetToughness.Val.Value = 1",
		"Sequence[1].Primitive.(game.CreateToken).Source.copy.AddSubtypes[0] = types.Nightmare",
	)
	// The gate must name the exile's own published result rather than some
	// other key, and the ability must be exactly the two instructions.
	paths := cardDefPaths(t, card)
	published := pathValue(t, paths, "Sequence[0].PublishResult")
	gated := pathValue(t, paths, "Sequence[1].ResultGate.Val.Key")
	if published != gated {
		t.Fatalf("exile publishes %s but the token gate reads %s, want the same key", published, gated)
	}
	assertCardPathsAbsent(t, card, "Modes[0].Sequence[2]", "TriggeredAbilities[1]")
}

// TestGenerateExecutableNightmareShepherdSourceParses keeps one renderer-level
// check on this card: the emitted Go must actually parse. That is a property of
// the rendered text and cannot be expressed as a CardDef assertion.
func TestGenerateExecutableNightmareShepherdSourceParses(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(nightmareShepherdCard(), "n")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "nightmare_shepherd.go", source, parser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
}
