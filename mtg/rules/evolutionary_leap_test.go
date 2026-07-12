package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestRevealUntilMatchToHandBottomsRemainder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Types: []types.Card{types.Creature},
	}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Land}}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Instant}}})

	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.RevealUntil{
		Player:                             game.ControllerReference(),
		Until:                              game.Selection{RequiredTypes: []types.Card{types.Creature}},
		Destination:                        zone.Hand,
		MatchToDestinationRestRandomBottom: true,
	}, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(creature) || g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatalf("hand = %v, want only matching creature", g.Players[game.Player1].Hand.All())
	}
	if g.Players[game.Player1].Library.Size() != 2 {
		t.Fatalf("library size = %d, want two remainder cards", g.Players[game.Player1].Library.Size())
	}
}
