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

func moreThanMeetsTheEyeBotDef() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Convert Bot",
			Types:     []types.Card{types.Artifact, types.Creature},
			ManaCost:  opt.Val(cost.Mana{cost.O(2)}),
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			AlternativeCosts: []cost.Alternative{{
				Label:    "More Than Meets the Eye",
				Mechanic: cost.AlternativeMechanicMoreThanMeetsTheEye,
				ManaCost: opt.Val(cost.Mana{cost.R, cost.W}),
			}},
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:      "Convert Bot Vehicle",
			Types:     []types.Card{types.Artifact},
			Subtypes:  []types.Sub{types.Vehicle},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
		}),
	}
}

// TestMoreThanMeetsTheEyeCastEntersConverted drives the real cast→resolve path to
// prove that casting a transforming double-faced card for its "More Than Meets
// the Eye" alternative cost makes the resulting permanent enter converted, as its
// back face (CR 712).
func TestMoreThanMeetsTheEyeCastEntersConverted(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, moreThanMeetsTheEyeBotDef())
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	agents := [game.NumPlayers]PlayerAgent{game.Player1: evokeCostAgent{costLabel: "More Than Meets the Eye"}}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(spellID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("More Than Meets the Eye cast failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Converted {
		t.Fatalf("stack object = %#v, want Converted spell", obj)
	}
	resolveStackWithTriggers(engine, g, agents)

	permanent := permanentForCard(g, spellID)
	if permanent == nil {
		t.Fatal("cast card did not enter the battlefield")
	}
	if permanent.Face != game.FaceBack || !permanent.Transformed {
		t.Fatalf("permanent face/transformed = %v/%v, want back/true", permanent.Face, permanent.Transformed)
	}
}

// TestNormalCastEntersFrontFace proves the same card cast for its normal mana
// cost enters as its front face, so only the More Than Meets the Eye alternative
// makes it enter converted.
func TestNormalCastEntersFrontFace(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, moreThanMeetsTheEyeBotDef())
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	agents := [game.NumPlayers]PlayerAgent{game.Player1: evokeCostAgent{costLabel: "Normal cost"}}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(spellID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("normal cast failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Converted {
		t.Fatalf("stack object = %#v, want non-converted spell", obj)
	}
	resolveStackWithTriggers(engine, g, agents)

	permanent := permanentForCard(g, spellID)
	if permanent == nil {
		t.Fatal("cast card did not enter the battlefield")
	}
	if permanent.Face != game.FaceFront || permanent.Transformed {
		t.Fatalf("permanent face/transformed = %v/%v, want front/false", permanent.Face, permanent.Transformed)
	}
}
