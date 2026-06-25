package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// evokeCostAgent casts the spell for a chosen cost-option label (the Evoke
// alternative or the normal cost) and otherwise orders triggered abilities in
// the order presented.
type evokeCostAgent struct {
	costLabel string
}

func (evokeCostAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a evokeCostAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoicePayment {
		for _, option := range request.Options {
			if option.Label == a.costLabel {
				return []int{option.Index}
			}
		}
	}
	indices := make([]int, 0, len(request.Options))
	for i := range request.Options {
		indices = append(indices, request.Options[i].Index)
	}
	return indices
}

func evokeTestCreature() *game.CardDef {
	pt := game.PT{Value: 1}
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Evoke Tester",
		Types:     []types.Card{types.Creature},
		ManaCost:  opt.Val(cost.Mana{cost.O(2)}),
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		AlternativeCosts: []cost.Alternative{{
			Label:    "Evoke",
			Mechanic: cost.AlternativeMechanicEvoke,
			ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		}},
		TriggeredAbilities: []game.TriggeredAbility{
			game.EvokeSacrificeTriggeredAbility(),
			{
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhen,
					Pattern: game.TriggerPattern{
						Event:  game.EventPermanentEnteredBattlefield,
						Source: game.TriggerSourceSelf,
					},
				},
				Content: (game.Mode{Sequence: []game.Instruction{{
					Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)},
				}}}).Ability(),
			},
		},
	}}
}

func resolveStackWithTriggers(engine *Engine, g *game.Game, agents [game.NumPlayers]PlayerAgent) {
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log)
	for {
		if _, ok := g.Stack.Peek(); !ok {
			break
		}
		engine.resolveTopOfStackWithChoices(g, agents, log)
		engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log)
	}
}

func TestEvokeCastSacrificesOnEntryAndStillResolvesETB(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	drawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	spellID := addCardToHand(g, game.Player1, evokeTestCreature())
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	agents := [game.NumPlayers]PlayerAgent{game.Player1: evokeCostAgent{costLabel: "Evoke"}}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(spellID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("evoke cast failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Evoked {
		t.Fatalf("stack object = %#v, want Evoked spell", obj)
	}
	resolveStackWithTriggers(engine, g, agents)

	if permanentForCard(g, spellID) != nil {
		t.Fatal("evoked creature stayed on the battlefield, want sacrificed")
	}
	if !g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("evoked creature was not sacrificed to the graveyard")
	}
	if !g.Players[game.Player1].Hand.Contains(drawn) {
		t.Fatal("enters-the-battlefield draw did not resolve for the evoked creature")
	}
}

func TestNormalCastDoesNotSacrifice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	drawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	spellID := addCardToHand(g, game.Player1, evokeTestCreature())
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	agents := [game.NumPlayers]PlayerAgent{game.Player1: evokeCostAgent{costLabel: "Normal cost"}}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(spellID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("normal cast failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Evoked {
		t.Fatalf("stack object = %#v, want non-evoked spell", obj)
	}
	resolveStackWithTriggers(engine, g, agents)

	if permanentForCard(g, spellID) == nil {
		t.Fatal("normally cast creature was sacrificed, want it on the battlefield")
	}
	if g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("normally cast creature went to the graveyard")
	}
	if !g.Players[game.Player1].Hand.Contains(drawn) {
		t.Fatal("enters-the-battlefield draw did not resolve for the normal creature")
	}
}
