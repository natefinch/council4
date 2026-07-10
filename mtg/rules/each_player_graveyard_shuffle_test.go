package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

func TestEachPlayerShufflesGraveyardIntoLibrary(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	var cards []struct {
		player game.PlayerID
		id     id.ID
	}
	for _, playerID := range []game.PlayerID{game.Player1, game.Player2} {
		cardID := g.IDGen.Next()
		g.CardInstances[cardID] = &game.CardInstance{
			ID:    cardID,
			Def:   vanillaCreature("Graveyard Card", 1, 1),
			Owner: playerID,
		}
		g.Players[playerID].Graveyard.Add(cardID)
		cards = append(cards, struct {
			player game.PlayerID
			id     id.ID
		}{player: playerID, id: cardID})
	}
	obj := &game.StackObject{Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.ShuffleGraveyardIntoLibrary{
		PlayerGroup: game.AllPlayersReference(),
	}, nil)

	for _, card := range cards {
		if g.Players[card.player].Graveyard.Contains(card.id) {
			t.Fatalf("player %v card remained in graveyard", card.player)
		}
		if !g.Players[card.player].Library.Contains(card.id) {
			t.Fatalf("player %v card not moved to library", card.player)
		}
	}
}
