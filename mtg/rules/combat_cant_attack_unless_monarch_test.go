package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// cantAttackUnlessDefenderIsMonarch builds a creature that can't attack unless
// the defending player is the monarch, matching the runtime shape produced by
// lowering "This creature can't attack unless defending player is the monarch."
// (Crown-Hunter Hireling).
func cantAttackUnlessDefenderIsMonarch(g *game.Game, controller game.PlayerID) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Crown Hunter",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:                    game.RuleEffectCantAttack,
				AffectedSource:          true,
				AttackDefenderIsMonarch: true,
			}},
		}},
	}})
}

func TestCantAttackUnlessDefenderIsMonarchGatesTargetByDesignation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := cantAttackUnlessDefenderIsMonarch(g, game.Player1)
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1

	// The defending player is not the monarch: the attacker can't attack them.
	if canAttackTarget(g, attacker, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("attacker may attack a defender who is not the monarch")
	}

	// Make the defending player the monarch: now the attack is legal.
	g.Players[game.Player2].IsMonarch = true
	if !canAttackTarget(g, attacker, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("attacker may not attack a defender who is the monarch")
	}
}

func TestCantAttackUnlessDefenderIsMonarchChecksDefenderNotAttacker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := cantAttackUnlessDefenderIsMonarch(g, game.Player1)
	// The attacker's own controller is the monarch; the defender is not.
	g.Players[game.Player1].IsMonarch = true
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1

	if canAttackTarget(g, attacker, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("attacker's own monarch designation satisfied a defender restriction")
	}
}

func TestCantAttackUnlessDefenderIsMonarchNilTargetIsLenient(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := cantAttackUnlessDefenderIsMonarch(g, game.Player1)
	g.Players[game.Player2].IsMonarch = true
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1

	// With a qualifying monarch defender on the board, the "can it attack at all"
	// check (nil target) must not forbid the creature from attacking.
	if !canAttackWith(g, attacker, game.Player1) {
		t.Fatal("attacker cannot attack at all despite a monarch defender")
	}
}
