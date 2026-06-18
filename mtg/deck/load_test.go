package deck_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/deck"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func legendaryCreature(name string, colors ...color.Color) *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(colors...),
		CardFace: game.CardFace{
			Name:       name,
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
		},
	}
}

func basicLand(name string, colors ...color.Color) *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(colors...),
		CardFace: game.CardFace{
			Name:       name,
			Supertypes: []types.Super{types.Basic},
			Types:      []types.Card{types.Land},
		},
	}
}

func nonbasicCreature(name string, colors ...color.Color) *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(colors...),
		CardFace: game.CardFace{
			Name:  name,
			Types: []types.Card{types.Creature},
		},
	}
}

func testRegistry() *cards.Registry {
	return cards.NewRegistry([]*game.CardDef{
		legendaryCreature("Test Commander", color.Green),
		basicLand("Forest", color.Green),
		nonbasicCreature("Llanowar Elves", color.Green),
	})
}

func validDecklist() *deck.Decklist {
	return &deck.Decklist{
		Commander: []deck.Entry{{Quantity: 1, Name: "Test Commander"}},
		Cards:     []deck.Entry{{Quantity: 99, Name: "Forest"}},
	}
}

func fourInputs(decklists ...*deck.Decklist) [game.NumPlayers]deck.PlayerInput {
	var inputs [game.NumPlayers]deck.PlayerInput
	for i := range inputs {
		dl := decklists[len(decklists)-1]
		if i < len(decklists) {
			dl = decklists[i]
		}
		inputs[i] = deck.PlayerInput{Name: fmt.Sprintf("P%d", i+1), Decklist: dl}
	}
	return inputs
}

func TestLoadCleanFourDecks(t *testing.T) {
	res := deck.Load(fourInputs(validDecklist()), game.Player1, testRegistry())

	if !res.OK() {
		t.Fatalf("expected OK load, got unresolved=%v legality=%v", res.Unresolved, res.Legality)
	}
	if res.UnderTest != game.Player1 {
		t.Errorf("UnderTest = %v, want Player1", res.UnderTest)
	}
	for i := range res.Configs {
		config := res.Configs[i]
		if config.Commander == nil || config.Commander.Name != "Test Commander" {
			t.Errorf("player %d commander = %v", i, config.Commander)
		}
		if len(config.Deck) != 99 {
			t.Errorf("player %d deck size = %d, want 99", i, len(config.Deck))
		}
	}
}

func TestLoadUnknownCard(t *testing.T) {
	illegal := &deck.Decklist{
		Commander: []deck.Entry{{Quantity: 1, Name: "Test Commander"}},
		Cards: []deck.Entry{
			{Quantity: 98, Name: "Forest"},
			{Quantity: 1, Name: "Definitely Not A Real Card"},
		},
	}
	res := deck.Load(fourInputs(illegal, validDecklist(), validDecklist(), validDecklist()), game.Player1, testRegistry())

	if res.OK() {
		t.Fatal("expected the load to report problems")
	}
	found := false
	for _, u := range res.Unresolved {
		if u.Player == game.Player1 && u.Name == "Definitely Not A Real Card" {
			found = true
		}
	}
	if !found {
		t.Errorf("Unresolved = %v, want the bogus card for Player1", res.Unresolved)
	}
}

func TestLoadIllegalDeckSurfacesLegality(t *testing.T) {
	// 97 Forests + 2 Llanowar Elves: deck size 99, but the duplicated nonbasic
	// violates the singleton rule.
	illegal := &deck.Decklist{
		Commander: []deck.Entry{{Quantity: 1, Name: "Test Commander"}},
		Cards: []deck.Entry{
			{Quantity: 97, Name: "Forest"},
			{Quantity: 2, Name: "Llanowar Elves"},
		},
	}
	res := deck.Load(fourInputs(illegal, validDecklist(), validDecklist(), validDecklist()), game.Player1, testRegistry())

	if res.OK() {
		t.Fatal("expected legality errors")
	}
	if len(res.Unresolved) != 0 {
		t.Errorf("unexpected unresolved cards: %v", res.Unresolved)
	}
	found := false
	for _, e := range res.Legality {
		if e.Player == game.Player1 && strings.Contains(e.Reason, "duplicate nonbasic") {
			found = true
		}
	}
	if !found {
		t.Errorf("Legality = %v, want a duplicate nonbasic error for Player1", res.Legality)
	}
}

func TestLoadMissingCommander(t *testing.T) {
	noCommander := &deck.Decklist{
		Cards: []deck.Entry{{Quantity: 99, Name: "Forest"}},
	}
	res := deck.Load(fourInputs(noCommander, validDecklist(), validDecklist(), validDecklist()), game.Player1, testRegistry())

	if res.Configs[game.Player1].Commander != nil {
		t.Errorf("Player1 commander = %v, want nil", res.Configs[game.Player1].Commander)
	}
	found := false
	for _, e := range res.Legality {
		if e.Player == game.Player1 && strings.Contains(e.Reason, "missing commander") {
			found = true
		}
	}
	if !found {
		t.Errorf("Legality = %v, want a missing-commander error for Player1", res.Legality)
	}
}

func TestParseThenLoad(t *testing.T) {
	text := "// Commander\n1 Test Commander\n// Deck\n99 Forest\n"
	dl, err := deck.Parse(strings.NewReader(text))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	res := deck.Load(fourInputs(dl), game.Player2, testRegistry())
	if !res.OK() {
		t.Fatalf("expected OK load, got unresolved=%v legality=%v", res.Unresolved, res.Legality)
	}
	if res.UnderTest != game.Player2 {
		t.Errorf("UnderTest = %v, want Player2", res.UnderTest)
	}
}
