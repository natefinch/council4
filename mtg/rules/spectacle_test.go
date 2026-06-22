package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func spectacleTestSpell() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:         "Spectacle Bolt",
		ManaCost:     opt.Val(cost.Mana{cost.O(2), cost.R}),
		Types:        []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.AbilityContent{}),
		AlternativeCosts: []cost.Alternative{{
			Label:     "Spectacle",
			ManaCost:  opt.Val(cost.Mana{cost.R}),
			Condition: cost.AlternativeConditionOpponentLostLifeThisTurn,
		}},
	}}
}

// TestSpectacleAlternativeCostRequiresOpponentLostLife verifies that the
// Spectacle alternative cost is only offered once an opponent has lost life this
// turn, and that the controller's own life loss does not enable it. The spell's
// normal cost ({2}{R}) is unpayable with a single Mountain, so the cast is legal
// only when the cheaper Spectacle cost ({R}) becomes available.
func TestSpectacleAlternativeCostRequiresOpponentLostLife(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, spectacleTestSpell())
	mountain := addBasicLandPermanent(g, game.Player1, types.Mountain)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	act := action.CastSpell(spellID, nil, 0, nil)

	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("spectacle cast was legal before any opponent lost life")
	}

	emitEvent(g, game.Event{Kind: game.EventLifeLost, Player: game.Player1, Amount: 5})
	if containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("controller's own life loss enabled the spectacle cost")
	}

	emitEvent(g, game.Event{Kind: game.EventLifeLost, Player: game.Player2, Amount: 3})
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("spectacle cast was not legal after an opponent lost life")
	}

	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("spectacle cast failed after an opponent lost life")
	}
	if !mountain.Tapped {
		t.Fatal("spectacle cost {R} was not paid from the Mountain")
	}
	if _, ok := g.Stack.Peek(); !ok {
		t.Fatal("spell was not put on the stack")
	}
}
