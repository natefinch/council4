package rules

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

func addCardToLibraryNamed(g *game.Game, player game.PlayerID, name string) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Sorcery}}},
		Owner: player,
	}
	g.Players[player].Library.Add(cardID)
	return cardID
}

func sortedIDs(ids []id.ID) []id.ID {
	out := append([]id.ID(nil), ids...)
	slices.Sort(out)
	return out
}

func TestDeterminizePreservesHandsAndDeckMultisets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	sorcery := func(n string) *game.CardDef {
		return &game.CardDef{CardFace: game.CardFace{Name: n, Types: []types.Card{types.Sorcery}}}
	}

	observerHand := []id.ID{
		addCardToHand(g, game.Player1, sorcery("Own-A")),
		addCardToHand(g, game.Player1, sorcery("Own-B")),
	}
	for range 5 {
		addCardToLibraryNamed(g, game.Player1, "Own-Lib")
	}

	addCardToHand(g, game.Player2, sorcery("Opp-Hand-1"))
	addCardToHand(g, game.Player2, sorcery("Opp-Hand-2"))
	for range 6 {
		addCardToLibraryNamed(g, game.Player2, "Opp-Lib")
	}

	opponentHandSize := g.Players[game.Player2].Hand.Size()
	opponentPool := sortedIDs(append(g.Players[game.Player2].Hand.All(), g.Players[game.Player2].Library.All()...))
	observerLibrarySize := g.Players[game.Player1].Library.Size()

	determinize(g, game.Player1, rand.New(rand.NewPCG(42, 42)))

	if got := sortedIDs(g.Players[game.Player1].Hand.All()); !slices.Equal(got, sortedIDs(observerHand)) {
		t.Fatalf("observer hand changed: got %v, want %v", got, observerHand)
	}
	if got := g.Players[game.Player1].Library.Size(); got != observerLibrarySize {
		t.Fatalf("observer library size = %d, want %d", got, observerLibrarySize)
	}
	if got := g.Players[game.Player2].Hand.Size(); got != opponentHandSize {
		t.Fatalf("opponent hand size = %d, want %d (preserved)", got, opponentHandSize)
	}
	gotPool := sortedIDs(append(g.Players[game.Player2].Hand.All(), g.Players[game.Player2].Library.All()...))
	if !slices.Equal(gotPool, opponentPool) {
		t.Fatalf("opponent deck multiset changed: got %v, want %v", gotPool, opponentPool)
	}
}
