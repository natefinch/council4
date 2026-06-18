package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestOpponentOrPlaneswalkerTargetLegality verifies the target spec lowered for
// "target opponent or planeswalker": an opponent player and any planeswalker are
// legal recipients, while a nonplaneswalker permanent and the caster are not.
func TestOpponentOrPlaneswalkerTargetLegality(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	opponentWalker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Foe Walker",
		Types:   []types.Card{types.Planeswalker},
		Loyalty: opt.Val(3)},
	})
	ownWalker := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Ally Walker",
		Types:   []types.Card{types.Planeswalker},
		Loyalty: opt.Val(3)},
	})
	creature := addCombatCreaturePermanent(g, game.Player2)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	spellID := addCardToHand(g, game.Player1, permanentTargetSpellWithSpecs([]game.TargetSpec{
		{
			MinTargets: 1,
			MaxTargets: 1,
			Constraint: "target opponent or planeswalker",
			Allow:      game.TargetAllowPlayer | game.TargetAllowPermanent,
			Predicate: game.TargetPredicate{
				Player:         game.PlayerOpponent,
				PermanentTypes: []types.Card{types.Planeswalker},
			},
		},
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if !containsAction(legal, action.CastSpell(spellID, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) {
		t.Fatal("opponent player was not a legal target")
	}
	if !containsAction(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(opponentWalker.ObjectID)}, 0, nil)) {
		t.Fatal("opponent-controlled planeswalker was not a legal target")
	}
	if !containsAction(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(ownWalker.ObjectID)}, 0, nil)) {
		t.Fatal("own planeswalker should be a legal target (any planeswalker)")
	}
	if containsAction(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0, nil)) {
		t.Fatal("nonplaneswalker creature must not be a legal target")
	}
	if containsAction(legal, action.CastSpell(spellID, []game.Target{game.PlayerTarget(game.Player1)}, 0, nil)) {
		t.Fatal("caster (not an opponent) must not be a legal player target")
	}
}
