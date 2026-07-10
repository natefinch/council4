package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func dashTestCreature() *game.CardDef {
	pt := game.PT{Value: 2}
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Dash Tester",
		Types:     []types.Card{types.Creature},
		ManaCost:  opt.Val(cost.Mana{cost.O(2)}),
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		AlternativeCosts: []cost.Alternative{{
			Label:    "Dash",
			Mechanic: cost.AlternativeMechanicDash,
			ManaCost: opt.Val(cost.Mana{cost.O(1)}),
		}},
		TriggeredAbilities: []game.TriggeredAbility{
			game.DashTriggeredAbility(),
		},
	}}
}

// TestDashCastGrantsHasteAndReturnsAtEndStep proves casting for the Dash
// alternative cost marks the spell dashed, grants the creature haste on entry,
// and returns it to its owner's hand at the beginning of the next end step.
func TestDashCastGrantsHasteAndReturnsAtEndStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, dashTestCreature())
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	agents := [game.NumPlayers]PlayerAgent{game.Player1: evokeCostAgent{costLabel: "Dash"}}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(spellID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("dash cast failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.Dashed {
		t.Fatalf("stack object = %#v, want Dashed spell", obj)
	}
	resolveStackWithTriggers(engine, g, agents)

	permanent := permanentForCard(g, spellID)
	if permanent == nil {
		t.Fatal("dashed creature is not on the battlefield after resolution")
	}
	if !hasKeyword(g, permanent, game.Haste) {
		t.Fatal("dashed creature did not gain haste")
	}

	g.Turn.Step = game.StepEnd
	emitBeginningOfStepEvent(g, game.StepEnd)
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	resolveStackWithTriggers(engine, g, agents)

	if permanentForCard(g, spellID) != nil {
		t.Fatal("dashed creature stayed on the battlefield past the end step, want returned to hand")
	}
	if !g.Players[game.Player1].Hand.Contains(spellID) {
		t.Fatal("dashed creature was not returned to its owner's hand at the end step")
	}
}

// TestNormalCastDoesNotDash proves casting for the normal mana cost leaves the
// spell undashed, grants no haste, and never returns the creature to hand.
func TestNormalCastDoesNotDash(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, dashTestCreature())
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	agents := [game.NumPlayers]PlayerAgent{game.Player1: evokeCostAgent{costLabel: "Normal cost"}}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(spellID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("normal cast failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Dashed {
		t.Fatalf("stack object = %#v, want non-dashed spell", obj)
	}
	resolveStackWithTriggers(engine, g, agents)

	permanent := permanentForCard(g, spellID)
	if permanent == nil {
		t.Fatal("normally cast creature is not on the battlefield after resolution")
	}
	if hasKeyword(g, permanent, game.Haste) {
		t.Fatal("normally cast creature gained haste, want summoning sick")
	}

	g.Turn.Step = game.StepEnd
	emitBeginningOfStepEvent(g, game.StepEnd)
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	resolveStackWithTriggers(engine, g, agents)

	if permanentForCard(g, spellID) == nil {
		t.Fatal("normally cast creature was returned to hand at the end step, want it to stay")
	}
	if g.Players[game.Player1].Hand.Contains(spellID) {
		t.Fatal("normally cast creature was returned to its owner's hand, want it on the battlefield")
	}
}

// TestDashReturnBindsToDashedObjectNotCard proves the delayed end-step return
// tracks the specific dashed object (CR 702.109a), not the card. When the
// dashed object leaves the battlefield and the same card re-enters as a new,
// undashed object before the end step (reanimation, blink, or a hard recast),
// the stale delayed trigger must no-op instead of wrongly bouncing the new
// object. Binding the return to SourceCardPermanentReference (by card identity)
// regresses this; binding it to SourcePermanentReference (by object identity)
// fixes it.
func TestDashReturnBindsToDashedObjectNotCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, dashTestCreature())
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	agents := [game.NumPlayers]PlayerAgent{game.Player1: evokeCostAgent{costLabel: "Dash"}}
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(spellID, nil, 0, nil), agents, &TurnLog{}) {
		t.Fatal("dash cast failed")
	}
	resolveStackWithTriggers(engine, g, agents)

	dashedObject := permanentForCard(g, spellID)
	if dashedObject == nil {
		t.Fatal("dashed creature is not on the battlefield after resolution")
	}
	dashedObjectID := dashedObject.ObjectID

	// The dashed object O1 leaves the battlefield to the graveyard, then the same
	// card re-enters as a distinct, undashed object O2 (as if reanimated) before
	// the end step.
	if !movePermanentToZone(g, dashedObject, zone.Graveyard) {
		t.Fatal("failed to move the dashed object off the battlefield")
	}
	reentered := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: spellID,
		Owner:          game.Player1,
		Controller:     game.Player1,
	}
	g.Battlefield = append(g.Battlefield, reentered)
	if reentered.ObjectID == dashedObjectID {
		t.Fatal("re-entered object reused the dashed object's ID; test cannot distinguish them")
	}

	g.Turn.Step = game.StepEnd
	emitBeginningOfStepEvent(g, game.StepEnd)
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	resolveStackWithTriggers(engine, g, agents)

	survivor := permanentForCard(g, spellID)
	if survivor == nil || survivor.ObjectID != reentered.ObjectID {
		t.Fatal("re-entered object was returned to hand; the dashed return must track the dashed object, not the card")
	}
	if g.Players[game.Player1].Hand.Contains(spellID) {
		t.Fatal("re-entered object was wrongly returned to its owner's hand at the end step")
	}
}
