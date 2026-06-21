package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestSearchToHandThenLoseLifeResolvesBoth proves the Search-then-life-loss
// sequence the cardgen backend emits for Grim Tutor ("Search your library for a
// card, put that card into your hand, then shuffle. You lose 3 life.") both
// moves the chosen card to its owner's hand and applies the trailing controller
// life loss.
func TestSearchToHandThenLoseLifeResolvesBoth(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	bear := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bear", Types: []types.Card{types.Creature}}})
	wolf := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Wolf", Types: []types.Card{types.Creature}}})
	before := g.Players[game.Player1].Life

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.Search{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
			Spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Hand,
			},
		}},
		{Primitive: game.LoseLife{
			Player: game.ControllerReference(),
			Amount: game.Fixed(3),
		}},
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: "Wolf"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(wolf) || g.Players[game.Player1].Library.Contains(wolf) {
		t.Fatal("search did not move the chosen card to hand")
	}
	if !g.Players[game.Player1].Library.Contains(bear) {
		t.Fatal("search moved an unchosen card out of the library")
	}
	if got := before - g.Players[game.Player1].Life; got != 3 {
		t.Fatalf("controller life loss = %d, want 3 from the trailing rider", got)
	}
}

// TestSearchToHandThenGainLifeResolvesBoth proves the trailing rider also models
// a fixed controller life gain — Environmental Sciences' "...then shuffle. You
// gain 2 life." — applied after the search resolves.
func TestSearchToHandThenGainLifeResolvesBoth(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	land := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest", Types: []types.Card{types.Land}}})
	before := g.Players[game.Player1].Life

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.Search{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
			Spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Hand,
			},
		}},
		{Primitive: game.GainLife{
			Player: game.ControllerReference(),
			Amount: game.Fixed(2),
		}},
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: "Forest"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(land) {
		t.Fatal("search did not move the chosen land to hand")
	}
	if got := g.Players[game.Player1].Life - before; got != 2 {
		t.Fatalf("controller life gain = %d, want 2 from the trailing rider", got)
	}
}
