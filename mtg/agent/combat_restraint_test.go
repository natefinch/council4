package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/rules"
)

func attackWith(attackers []*game.Permanent, defender game.PlayerID) action.Action {
	declarations := make([]game.AttackDeclaration, 0, len(attackers))
	for _, attacker := range attackers {
		declarations = append(declarations, game.AttackDeclaration{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: defender},
		})
	}
	return action.DeclareAttackers(declarations)
}

func TestHoldsBackWhenAttackingRisksLethalCrackback(t *testing.T) {
	// A 5/5 into a 6/1 is a fine attack in isolation, but committing my only
	// blocker while at 5 life lets the 6/1 swing back for lethal. The same attack
	// at 40 life is safe.
	build := func(life int) (rules.PlayerObservation, action.Action) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		g.Players[game.Player1].Life = life
		mine := addObservedPermanent(g, game.Player1, creatureCardDef("Knight", 5, 5))
		addObservedPermanent(g, game.Player2, creatureCardDef("Spear", 6, 1))
		return rules.NewObservation(g, game.Player1), attackWith([]*game.Permanent{mine}, game.Player2)
	}

	strategy := GenericStrategy{}
	lowObs, lowAttack := build(5)
	highObs, highAttack := build(40)

	lowScore := strategy.ScoreAction(lowObs, lowAttack)
	highScore := strategy.ScoreAction(highObs, highAttack)

	if lowScore >= scorePass {
		t.Fatalf("attacking into lethal crackback scored %v, want below not attacking %v", lowScore, scorePass)
	}
	if highScore <= scorePass {
		t.Fatalf("the same attack at 40 life scored %v, want a safe, positive attack", highScore)
	}
}

func TestAttacksFreelyWhenNoCrackback(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].Life = 12
	mine := addObservedPermanent(g, game.Player1, creatureCardDef("Knight", 5, 5))
	// Opponent has no creatures, so there is no crackback and no penalty.
	obs := rules.NewObservation(g, game.Player1)

	strategy := GenericStrategy{}
	if score := strategy.ScoreAction(obs, attackWith([]*game.Permanent{mine}, game.Player2)); score <= scorePass {
		t.Fatalf("a safe attack scored %v, want positive", score)
	}
}

func TestKeepsABlockerBackInsteadOfAllOutAttack(t *testing.T) {
	// At 5 life against a 6/1: swinging with both 5/5s leaves no blocker and dies
	// to the crackback, but attacking with one and keeping one back blocks the
	// 6/1 and survives, so the restrained attack must score higher.
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Players[game.Player1].Life = 5
	first := addObservedPermanent(g, game.Player1, creatureCardDef("Knight", 5, 5))
	second := addObservedPermanent(g, game.Player1, creatureCardDef("Squire", 5, 5))
	addObservedPermanent(g, game.Player2, creatureCardDef("Spear", 6, 1))
	obs := rules.NewObservation(g, game.Player1)

	strategy := GenericStrategy{}
	allOut := strategy.ScoreAction(obs, attackWith([]*game.Permanent{first, second}, game.Player2))
	restrained := strategy.ScoreAction(obs, attackWith([]*game.Permanent{first}, game.Player2))

	if restrained <= allOut {
		t.Fatalf("attacking with one and holding a blocker scored %v, want above the all-out attack %v", restrained, allOut)
	}
}
