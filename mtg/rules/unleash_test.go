package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func enterUnleashCreature(t *testing.T, g *game.Game, engine *Engine, agents [game.NumPlayers]PlayerAgent) *game.Permanent {
	t.Helper()
	def := &game.CardDef{CardFace: game.CardFace{
		Name:            "Unleash Ogre",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 2}),
		Toughness:       opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{game.UnleashStaticBody},
	}}
	cardID := addCardToHand(g, game.Player1, def)
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatal("card instance not found")
	}
	g.Players[game.Player1].Hand.Remove(cardID)
	permanent, ok := createCardPermanentWithChoices(engine, g, card, game.Player1, zone.Hand, agents, &TurnLog{})
	if !ok {
		t.Fatal("createCardPermanentWithChoices() = false, want true")
	}
	return permanent
}

// TestUnleashEntersWithCounterCantBlock verifies that an unleash creature whose
// controller chooses the +1/+1 counter enters with one counter and can't block
// while it has the counter (CR 702.86).
func TestUnleashEntersWithCounterCantBlock(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}
	permanent := enterUnleashCreature(t, g, engine, agents)
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("unleash counter choice: +1/+1 counters = %d, want 1", got)
	}
	if canBlockWith(g, permanent, game.Player1) {
		t.Fatal("unleash creature with a +1/+1 counter must not be able to block")
	}
}

// TestUnleashDeclineCounterCanBlock verifies that declining the unleash counter
// leaves the creature with no counter and able to block.
func TestUnleashDeclineCounterCanBlock(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}
	permanent := enterUnleashCreature(t, g, engine, agents)
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("unleash decline: +1/+1 counters = %d, want 0", got)
	}
	if !canBlockWith(g, permanent, game.Player1) {
		t.Fatal("unleash creature without a counter must be able to block")
	}
}
