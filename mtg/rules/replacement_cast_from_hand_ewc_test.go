package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func castFromHandEntersWithCounterCreature() *game.CardDef {
	def := creatureSpellDef("Hand-Cast Spirit", types.Spirit)
	def.ManaCost = opt.Val(cost.Mana{cost.O(1)})
	def.Power = opt.Val(game.PT{Value: 1})
	def.Toughness = opt.Val(game.PT{Value: 1})
	def.ReplacementAbilities = []game.ReplacementAbility{
		game.EntersWithCountersIfReplacement(
			"Hand-Cast Spirit enters with a divinity counter on it if you cast it from your hand.",
			&game.Condition{EventPermanentWasCastFromControllerHand: true},
			game.CounterPlacement{Kind: counter.Divinity, Amount: 1},
		),
	}
	return def
}

// TestCastFromHandEntersWithCounterApplies verifies the condition succeeds when
// the permanent's spell was cast by its controller from that player's hand.
func TestCastFromHandEntersWithCounterApplies(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	cardID := addCardToHand(g, game.Player1, castFromHandEntersWithCounterCreature())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("cast action failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent, ok := findPermanentByCardID(g, cardID)
	if !ok {
		t.Fatal("cast creature did not enter")
	}
	if got := permanent.Counters.Get(counter.Divinity); got != 1 {
		t.Fatalf("divinity counters = %d, want 1", got)
	}
}

// TestCastFromHandEntryConditionFailsClosed verifies the typed condition rejects
// non-casts, casts from another zone, and casts by another player.
func TestCastFromHandEntryConditionFailsClosed(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatPermanent(g, game.Player1, castFromHandEntersWithCounterCreature())
	events := []game.Event{
		{
			Kind:        game.EventPermanentEnteredBattlefield,
			PermanentID: permanent.ObjectID,
			Controller:  game.Player1,
		},
		{
			Kind:                   game.EventPermanentEnteredBattlefield,
			PermanentID:            permanent.ObjectID,
			Controller:             game.Player1,
			EnterWasCast:           true,
			EnterHasCastController: true,
			EnterCastController:    game.Player1,
			EnterCastFromZone:      zone.Graveyard,
		},
		{
			Kind:                   game.EventPermanentEnteredBattlefield,
			PermanentID:            permanent.ObjectID,
			Controller:             game.Player1,
			EnterWasCast:           true,
			EnterHasCastController: true,
			EnterCastController:    game.Player2,
			EnterCastFromZone:      zone.Hand,
		},
	}
	condition := opt.Val(game.Condition{EventPermanentWasCastFromControllerHand: true})
	for i := range events {
		if conditionSatisfied(g, conditionContext{
			controller: game.Player1,
			source:     permanent,
			event:      &events[i],
		}, condition) {
			t.Fatalf("cast-from-hand condition succeeded for invalid event %d: %#v", i, events[i])
		}
	}
}
