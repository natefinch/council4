package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestFusedTriggerFiresOnEnterAndOnEvent verifies the runtime semantics that a
// fused "When ~ enters and whenever an opponent draws a card, <effect>" ability
// relies on: the parser splits it into two independent triggered abilities, one
// scoped to the source entering the battlefield and one to the joined event.
// Each must fire on its own event and not on the other's.
func TestFusedTriggerFiresOnEnterAndOnEvent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	draw := game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
	}}}.Ability()
	pt := game.PT{Value: 1}
	def := &game.CardDef{CardFace: game.CardFace{
		Name:      "Fused Tester",
		Types:     []types.Card{types.Creature},
		ManaCost:  greenCost(),
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		TriggeredAbilities: []game.TriggeredAbility{
			{
				Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
					Event:  game.EventPermanentEnteredBattlefield,
					Source: game.TriggerSourceSelf,
				}},
				Content: draw,
			},
			{
				Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
					Event:  game.EventCardDrawn,
					Player: game.TriggerPlayerOpponent,
				}},
				Content: draw,
			},
		},
	}}
	source := addCombatPermanent(g, game.Player1, def)
	card, ok := g.GetCardInstance(source.CardInstanceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	enterPattern := &card.Def.TriggeredAbilities[0].Trigger.Pattern
	drawPattern := &card.Def.TriggeredAbilities[1].Trigger.Pattern

	enterEvent := game.Event{Kind: game.EventPermanentEnteredBattlefield, Controller: game.Player1, PermanentID: source.ObjectID}
	opponentDraw := game.Event{Kind: game.EventCardDrawn, Player: game.Player2}
	controllerDraw := game.Event{Kind: game.EventCardDrawn, Player: game.Player1}

	if !triggerMatchesEvent(g, source, enterPattern, enterEvent) {
		t.Fatal("enter constituent did not fire on the source entering")
	}
	if triggerMatchesEvent(g, source, enterPattern, opponentDraw) {
		t.Fatal("enter constituent fired on a draw event")
	}
	if !triggerMatchesEvent(g, source, drawPattern, opponentDraw) {
		t.Fatal("draw constituent did not fire on an opponent draw")
	}
	if triggerMatchesEvent(g, source, drawPattern, controllerDraw) {
		t.Fatal("opponent-scoped draw constituent fired on the controller's draw")
	}
	if triggerMatchesEvent(g, source, drawPattern, enterEvent) {
		t.Fatal("draw constituent fired on the enter event")
	}
}
