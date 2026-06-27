package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

// selfCastTriggerCard is an instant whose own "When you cast this spell, draw a
// card" triggered ability fires from the stack as the spell is cast, mirroring
// the Nulldrifter shape (a self-cast spell trigger with a target-free body).
func selfCastTriggerCard() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Self Cast Trigger Card",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Instant},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhen,
				Pattern: game.TriggerPattern{
					Event:       game.EventSpellCast,
					Controller:  game.TriggerControllerYou,
					Source:      game.TriggerSourceSelf,
					SelfWasCast: true,
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

func TestCastSelfTriggerFiresOnCast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	drawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	spellID := addCardToHand(g, game.Player1, selfCastTriggerCard())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction() = false, want true for casting the spell")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self-cast trigger did not fire on cast")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackTriggeredAbility || obj.SourceCardID != spellID {
		t.Fatalf("top of stack = %+v, want self-cast triggered ability sourced from the cast spell", obj)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(drawn) {
		t.Fatalf("hand = %v, want the self-cast trigger to draw a card", g.Players[game.Player1].Hand)
	}
}

func TestCastSelfTriggerDoesNotFireForOtherSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	// The self-cast card sits in the graveyard; casting a different spell must
	// not fire its self-source cast trigger.
	addCardToGraveyard(g, game.Player1, selfCastTriggerCard())
	otherID := addCardToHand(g, game.Player1, greenInstant())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(otherID, nil, 0, nil)) {
		t.Fatal("applyAction() = false, want true for casting a different spell")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("self-cast trigger fired for a different cast spell")
	}
}
