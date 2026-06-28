package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// libraryGraveyardTutorInstruction builds the optional library-and-graveyard
// named-tutor Search instruction the dual-zone lowerer produces: search both the
// library and graveyard for a card with the given name, reveal it, and put it
// into the controller's hand.
func libraryGraveyardTutorInstruction(name string) game.Instruction {
	return game.Instruction{
		Optional: true,
		Primitive: game.Search{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
			Spec: game.SearchSpec{
				SourceZone:    zone.Library,
				Destination:   zone.Hand,
				Name:          name,
				Reveal:        true,
				AlsoGraveyard: true,
			},
		},
	}
}

func tutorTarget(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Creature}}}
}

// TestLibraryGraveyardTutorFindsFromLibrary verifies that accepting the optional
// dual-zone tutor and choosing the named card in the library moves it to hand.
func TestLibraryGraveyardTutorFindsFromLibrary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCardToLibrary(g, game.Player1, tutorTarget("Teferi, Timebender"))
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bear", Types: []types.Card{types.Creature}}})
	addInstructionSpellToStack(g, []game.Instruction{libraryGraveyardTutorInstruction("Teferi, Timebender")})
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalSearchAgent{accept: true, wanted: "Teferi, Timebender"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(target) {
		t.Fatal("named card from library was not moved to hand")
	}
	if g.Players[game.Player1].Library.Contains(target) {
		t.Fatal("named card was left in the library")
	}
}

// TestLibraryGraveyardTutorFindsFromGraveyard verifies the tutor also reaches the
// graveyard: with the named card only in the graveyard, accepting moves it to
// hand.
func TestLibraryGraveyardTutorFindsFromGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bear", Types: []types.Card{types.Creature}}})
	target := addCardToGraveyard(g, game.Player1, tutorTarget("Teferi, Timebender"))
	addInstructionSpellToStack(g, []game.Instruction{libraryGraveyardTutorInstruction("Teferi, Timebender")})
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalSearchAgent{accept: true, wanted: "Teferi, Timebender"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(target) {
		t.Fatal("named card from graveyard was not moved to hand")
	}
	if g.Players[game.Player1].Graveyard.Contains(target) {
		t.Fatal("named card was left in the graveyard")
	}
}

// TestLibraryGraveyardTutorDeclineLeavesZones verifies that declining the
// optional tutor leaves both the library and graveyard copies untouched.
func TestLibraryGraveyardTutorDeclineLeavesZones(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	libCard := addCardToLibrary(g, game.Player1, tutorTarget("Teferi, Timebender"))
	graveCard := addCardToGraveyard(g, game.Player1, tutorTarget("Teferi, Timebender"))
	addInstructionSpellToStack(g, []game.Instruction{libraryGraveyardTutorInstruction("Teferi, Timebender")})
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalSearchAgent{accept: false, wanted: "Teferi, Timebender"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Library.Contains(libCard) {
		t.Fatal("declining the tutor still moved the library copy")
	}
	if !g.Players[game.Player1].Graveyard.Contains(graveCard) {
		t.Fatal("declining the tutor still moved the graveyard copy")
	}
	if g.Players[game.Player1].Hand.Contains(libCard) || g.Players[game.Player1].Hand.Contains(graveCard) {
		t.Fatal("declining the tutor still added a card to hand")
	}
}
