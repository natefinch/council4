package rules

import (
	"testing"

	cardss "github.com/natefinch/council4/mtg/cards/s"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestSnuffOutFreeAlternativeCostPaysLifeWithSwamp proves the "free spell"
// alternative cost generated for Snuff Out: "If you control a Swamp, you may pay
// 4 life rather than pay this spell's mana cost." A single Swamp cannot pay the
// {3}{B} mana cost, so the only way to cast is the pay-4-life alternative, which
// the condition gates on controlling a Swamp. The cast must deduct 4 life,
// leave the Swamp untapped (mana was not paid), and destroy the nonblack target
// when it resolves.
func TestSnuffOutFreeAlternativeCostPaysLifeWithSwamp(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, cardss.SnuffOut())
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Grizzly Bears",
		Types:     []types.Card{types.Creature},
		Colors:    []color.Color{color.Green},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.CastSpell(spellID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("Snuff Out cast was legal without controlling a Swamp")
	}

	swamp := addBasicLandPermanent(g, game.Player1, types.Swamp)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("Snuff Out cast was not legal while controlling a Swamp")
	}

	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("applyAction(cast Snuff Out via pay 4 life) = false, want true")
	}
	if got := g.Players[game.Player1].Life; got != 36 {
		t.Fatalf("caster life = %d, want 36 after paying 4 life", got)
	}
	if swamp.Tapped {
		t.Fatal("Swamp was tapped; the mana cost must not be paid for the free alternative cost")
	}
	if _, ok := g.Stack.Peek(); !ok {
		t.Fatal("Snuff Out was not put on the stack")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, target.ObjectID); ok {
		t.Fatal("target creature was not destroyed")
	}
	if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("destroyed creature did not move to its owner's graveyard")
	}
}

// TestFreeAlternativeCostYourTurnGate covers the "If it's your turn," gate of
// the free-spell alternative cost family (e.g. Mine Collapse) at runtime: the
// alternative sacrifice cost is offered only while its controller is the active
// player. The {4}{R} mana cost is unpayable from a single Mountain, so the cast
// is legal only when the your-turn alternative is available.
func TestFreeAlternativeCostYourTurnGate(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spell := &game.CardDef{CardFace: game.CardFace{
		Name:         "Turn Gated Bolt",
		ManaCost:     opt.Val(cost.Mana{cost.O(4), cost.R}),
		Types:        []types.Card{types.Instant},
		SpellAbility: opt.Val(game.AbilityContent{}),
		AlternativeCosts: []cost.Alternative{{
			Label: "Sacrifice a Mountain",
			AdditionalCosts: []cost.Additional{{
				Kind:        cost.AdditionalSacrifice,
				Text:        "sacrifice a Mountain",
				Amount:      1,
				SubtypesAny: cost.SubtypeSet{types.Mountain},
			}},
			Condition: cost.AlternativeConditionYourTurn,
		}},
	}}
	spellID := addCardToHand(g, game.Player1, spell)
	addBasicLandPermanent(g, game.Player1, types.Mountain)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.CastSpell(spellID, nil, 0, nil)

	g.Turn.ActivePlayer = game.Player2
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("your-turn free cost was offered on the opponent's turn")
	}

	g.Turn.ActivePlayer = game.Player1
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("your-turn free cost was not offered on the controller's own turn")
	}
}
