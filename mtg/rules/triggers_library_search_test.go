package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func librarySearchTriggerInstructions() []game.Instruction {
	return []game.Instruction{{Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}}}
}

func resolveLibrarySearch(t *testing.T, g *game.Game, engine *Engine, searcher game.PlayerID) {
	t.Helper()
	addCardToLibrary(g, searcher, &game.CardDef{CardFace: game.CardFace{
		Name: "Bear", Types: []types.Card{types.Creature}}})
	addEffectSpellToStack(g, searcher, game.Search{
		Amount: game.Fixed(1),
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:  zone.Library,
			Destination: zone.Hand,
			CardType:    opt.Val(types.Creature),
		},
	}, nil)
	var agents [game.NumPlayers]PlayerAgent
	agents[searcher] = &searchByNameAgent{wanted: "Bear"}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
}

func TestLibrarySearchTriggerFiresWhenOpponentSearches(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Archivist of Oghma: "Whenever an opponent searches their library, ...",
	// controlled by Player2 so Player1 is the opponent who searches.
	addTriggeredPermanent(g, game.Player2, &game.TriggerPattern{
		Event:  game.EventLibrarySearched,
		Player: game.TriggerPlayerOpponent,
	}, librarySearchTriggerInstructions(), nil)

	resolveLibrarySearch(t, g, engine, game.Player1)

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("opponent-search trigger did not fire when an opponent searched their library")
	}
}

func TestLibrarySearchOpponentTriggerSkipsControllerSearch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// The same opponent-scoped trigger must not fire when its own controller
	// searches their library.
	addTriggeredPermanent(g, game.Player2, &game.TriggerPattern{
		Event:  game.EventLibrarySearched,
		Player: game.TriggerPlayerOpponent,
	}, librarySearchTriggerInstructions(), nil)

	resolveLibrarySearch(t, g, engine, game.Player2)

	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("opponent-search trigger fired when its controller searched their library")
	}
}
