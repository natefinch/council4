package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func unearthSource() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Unearth Source",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Zombie},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		ActivatedAbilities: []game.ActivatedAbility{
			game.UnearthActivatedAbility(cost.Mana{cost.O(0)}),
		},
	}}
}

func TestUnearthReturnsSourceToBattlefieldWithHaste(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	cardID := addCardToGraveyard(g, game.Player1, unearthSource())

	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	act := action.ActivateAbility(cardID, 0, nil, 0)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("unearth ability was not legal from the graveyard")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("unearth activation failed")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("unearth source remained in graveyard after resolution")
	}
	var permanent *game.Permanent
	for _, p := range g.Battlefield {
		if p.CardInstanceID == cardID {
			permanent = p
			break
		}
	}
	if permanent == nil {
		t.Fatal("unearth source did not return to the battlefield")
	}
	if permanent.Controller != game.Player1 {
		t.Fatalf("unearthed permanent controller = %v; want %v", permanent.Controller, game.Player1)
	}
	if !hasKeyword(g, permanent, game.Haste) {
		t.Fatal("unearthed permanent did not gain haste")
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %d; want 1 end-step exile", len(g.DelayedTriggers))
	}
	if g.DelayedTriggers[0].Timing != game.DelayedAtBeginningOfNextEndStep {
		t.Fatalf("delayed trigger timing = %v; want next end step", g.DelayedTriggers[0].Timing)
	}
}
