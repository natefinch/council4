package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func basicIslandDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Island",
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
		Subtypes:   []types.Sub{types.Island},
	}}
}

// TestEachPlayerSearchBasicLandToBattlefield proves the each-player (symmetric)
// library search: every player searches their OWN library for a basic land and
// puts it onto the battlefield under their own control. Player1 fetches their
// Forest, Player2 fetches their Island; neither finds the other's land because
// each searches only their own library.
func TestEachPlayerSearchBasicLandToBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forest := addCardToLibrary(g, game.Player1, basicForestDef())
	island := addCardToLibrary(g, game.Player2, basicIslandDef())
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount:      game.Fixed(1),
		PlayerGroup: game.AllPlayersReference(),
		Spec: game.SearchSpec{
			SourceZone:  zone.Library,
			Destination: zone.Battlefield,
			CardType:    opt.Val(types.Land),
			Supertype:   opt.Val(types.Basic),
		},
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: optionalSearchAgent{accept: true, wanted: "Forest"},
		game.Player2: optionalSearchAgent{accept: true, wanted: "Island"},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	forestPermanent := permanentForCard(g, forest)
	if forestPermanent == nil {
		t.Fatal("Player1's basic land did not enter the battlefield")
	}
	if forestPermanent.Controller != game.Player1 {
		t.Fatalf("Player1's fetched land controller = %v, want Player1", forestPermanent.Controller)
	}
	if g.Players[game.Player1].Library.Contains(forest) {
		t.Fatal("Player1's basic land was not removed from their library")
	}

	islandPermanent := permanentForCard(g, island)
	if islandPermanent == nil {
		t.Fatal("Player2's basic land did not enter the battlefield")
	}
	if islandPermanent.Controller != game.Player2 {
		t.Fatalf("Player2's fetched land controller = %v, want Player2", islandPermanent.Controller)
	}
	if g.Players[game.Player2].Library.Contains(island) {
		t.Fatal("Player2's basic land was not removed from their library")
	}
}
