package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestDynamicCardsDrawnThisTurnCountsControllerDraws covers Thundering Djinn's
// "equal to the number of cards you've drawn this turn.": a
// DynamicAmountCardsDrawnThisTurn counts only the controller's draws this turn
// and applies the amount's multiplier; an opponent's draws are ignored.
func TestDynamicCardsDrawnThisTurnCountsControllerDraws(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := &game.StackObject{Controller: game.Player1}
	amount := game.DynamicAmount{Kind: game.DynamicAmountCardsDrawnThisTurn}

	if got := dynamicAmountValue(g, obj, game.Player1, amount); got != 0 {
		t.Fatalf("cards drawn this turn = %d, want 0 before any draw", got)
	}

	emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player1})
	emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player1})
	emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player2})

	if got := dynamicAmountValue(g, obj, game.Player1, amount); got != 2 {
		t.Fatalf("cards drawn this turn = %d, want 2 (controller's two draws only)", got)
	}

	withMultiplier := game.DynamicAmount{Kind: game.DynamicAmountCardsDrawnThisTurn, Multiplier: 3}
	if got := dynamicAmountValue(g, obj, game.Player1, withMultiplier); got != 6 {
		t.Fatalf("thrice cards drawn this turn = %d, want 6", got)
	}
}

// TestDynamicStarPowerCardsDrawnThisTurn covers Duelist of the Mind's
// characteristic-defining power ("Duelist of the Mind's power is equal to the
// number of cards you've drawn this turn."): the power tracks the controller's
// live draw count this turn while the printed toughness stands.
func TestDynamicStarPowerCardsDrawnThisTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:         "Duelist of the Mind",
		Types:        []types.Card{types.Creature},
		Power:        opt.Val(game.PT{IsStar: true}),
		Toughness:    opt.Val(game.PT{Value: 3}),
		DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCardsDrawnThisTurn}),
	}})

	if got := effectivePower(g, creature); got != 0 {
		t.Fatalf("effective power = %d, want 0 before any draw", got)
	}

	emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player1})
	emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player2})

	if got := effectivePower(g, creature); got != 1 {
		t.Fatalf("effective power = %d, want 1 (controller's single draw)", got)
	}
	if toughness, ok := effectiveToughness(g, creature); !ok || toughness != 3 {
		t.Fatalf("effective toughness = %d (ok=%v), want printed 3", toughness, ok)
	}

	emitEvent(g, game.Event{Kind: game.EventCardDrawn, Player: game.Player1})
	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("effective power after extra draw = %d, want 2", got)
	}
}
