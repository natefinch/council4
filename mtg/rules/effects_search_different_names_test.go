package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// distinctNamesSearchSpec is the typed search a "with different names" multi-card
// tutor lowers to: up to three creature cards with different names, into hand,
// then shuffle.
func distinctNamesSearchSpec() game.SearchSpec {
	return game.SearchSpec{
		SourceZone:     zone.Library,
		Destination:    zone.Hand,
		DifferentNames: true,
		Reveal:         true,
		Filter:         game.Selection{RequiredTypes: []types.Card{types.Creature}},
	}
}

func creatureCard(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Creature}}}
}

// TestDifferentNamesSearchPicksDistinctNames verifies the staged choice only
// offers cards whose name is not already chosen, so two distinct-name creatures
// are both found and a duplicate-name copy is never assembled.
func TestDifferentNamesSearchPicksDistinctNames(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	bear := addCardToLibrary(g, game.Player1, creatureCard("Bear"))
	addCardToLibrary(g, game.Player1, creatureCard("Bear"))
	wolf := addCardToLibrary(g, game.Player1, creatureCard("Wolf"))
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(2),
		Player: game.ControllerReference(),
		Spec:   distinctNamesSearchSpec(),
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: newCorrelatedSearchAgent("Bear", "Wolf")}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	hand := g.Players[game.Player1].Hand.All()
	if len(hand) != 2 {
		t.Fatalf("expected exactly two distinct-name creatures in hand, got %v", hand)
	}
	names := map[string]int{}
	for _, c := range hand {
		if card, ok := g.GetCardInstance(c); ok {
			names[card.Def.Name]++
		}
	}
	if names["Bear"] != 1 || names["Wolf"] != 1 {
		t.Fatalf("expected one Bear and one Wolf, got %v", names)
	}
	_, _ = bear, wolf
}

// TestDifferentNamesSearchFindsOneWhenOnlyDuplicates verifies that when only
// same-named cards remain after the first pick, the search finds just one card
// rather than two copies of the same name.
func TestDifferentNamesSearchFindsOneWhenOnlyDuplicates(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	bearA := addCardToLibrary(g, game.Player1, creatureCard("Bear"))
	addCardToLibrary(g, game.Player1, creatureCard("Bear"))
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(2),
		Player: game.ControllerReference(),
		Spec:   distinctNamesSearchSpec(),
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: newCorrelatedSearchAgent("Bear")}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	count := 0
	for _, c := range g.Players[game.Player1].Hand.All() {
		if card, ok := g.GetCardInstance(c); ok && card.Def.Name == "Bear" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one Bear in hand, got %d", count)
	}
	_ = bearA
}
