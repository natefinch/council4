package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestConditionalDestinationPlaceMandatoryThenCannotBeDeclined(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	plains := addCardToLibrary(g, game.Player1, plainsCard())
	addLandPermanent(g, game.Player2, "Opp Land 1")
	addLandPermanent(g, game.Player2, "Opp Land 2")
	sequence := scholarSequence()
	place, ok := sequence[1].Primitive.(game.ConditionalDestinationPlace)
	if !ok {
		t.Fatalf("instruction 1 = %T, want ConditionalDestinationPlace", sequence[1].Primitive)
	}
	place.ThenMandatory = true
	sequence[1].Primitive = place
	addInstructionSpellToStack(g, sequence)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: conditionalDestinationAgent{wanted: "Plains", acceptPut: false}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player1].Library.Contains(plains) || g.Players[game.Player1].Hand.Contains(plains) {
		t.Fatal("mandatory matching card did not leave the library for the battlefield")
	}
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == plains {
			if !permanent.Tapped {
				t.Fatal("mandatory battlefield card should enter tapped")
			}
			return
		}
	}
	t.Fatal("mandatory matching card did not enter the battlefield")
}
