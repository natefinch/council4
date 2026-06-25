package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// selfCycleTriggerCard is a cycling card whose own "When you cycle this card"
// triggered ability draws a card, mirroring the Magmakin Artillerist shape
// (Cycling plus a self-source cycle trigger) but with a target-free body.
func selfCycleTriggerCard() *game.CardDef {
	pt := game.PT{Value: 2}
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Self Cycle Trigger Card",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		ActivatedAbilities: []game.ActivatedAbility{
			game.CyclingActivatedAbility(cost.Mana{cost.O(1)}),
		},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhen,
				Pattern: game.TriggerPattern{
					Event:  game.EventCycled,
					Player: game.TriggerPlayerYou,
					Source: game.TriggerSourceSelf,
				},
			},
			Content: game.Mode{
				Sequence: []game.Instruction{{
					Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
				}},
			}.Ability(),
		}},
	}}
}

func TestCycleSelfTriggerFiresFromGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cycleDrawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Cycle Drawn"}})
	triggerDrawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Trigger Drawn"}})
	cardID := addCardToHand(g, game.Player1, selfCycleTriggerCard())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(cardID, 0, nil, 0)) {
		t.Fatal("applyAction() = false, want true for cycling")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("cycled card was not discarded to graveyard")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self-cycle trigger did not fire from the graveyard")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackTriggeredAbility || obj.SourceCardID != cardID {
		t.Fatalf("top of stack = %+v, want self-cycle triggered ability sourced from cycled card", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	// Resolve the cycling activated ability's own draw too.
	if obj, ok := g.Stack.Peek(); ok && obj.Kind == game.StackActivatedAbility {
		engine.resolveTopOfStack(g, &TurnLog{})
	}
	hand := g.Players[game.Player1].Hand
	if !hand.Contains(cycleDrawn) || !hand.Contains(triggerDrawn) {
		t.Fatalf("hand = %v, want both cycling draw and self-trigger draw", hand)
	}
}

func TestCycleSelfTriggerDoesNotFireForOtherCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Cycle Drawn"}})
	// The self-cycle card sits in the graveyard; cycling a different card must
	// not fire its self-source trigger.
	addCardToGraveyard(g, game.Player1, selfCycleTriggerCard())
	otherID := addCardToHand(g, game.Player1, cyclingCard())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(otherID, 0, nil, 0)) {
		t.Fatal("applyAction() = false, want true for cycling other card")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self-cycle trigger fired for a different cycled card")
	}
}
