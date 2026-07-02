package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// cantAttackUnlessDefenderControlsIsland builds a creature that can't attack
// unless the defending player controls an Island, matching the runtime shape
// produced by lowering "This creature can't attack unless defending player
// controls an Island."
func cantAttackUnlessDefenderControlsIsland(g *game.Game, controller game.PlayerID) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Conditional Attacker",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:                            game.RuleEffectCantAttack,
				AffectedSource:                  true,
				AttackDefenderControlsSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Island")}},
			}},
		}},
	}})
}

func addIslandPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:     "Island",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Sub("Island")},
	}})
}

func TestCantAttackUnlessDefenderControlsIslandGatesTargetByBoard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := cantAttackUnlessDefenderControlsIsland(g, game.Player1)
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1

	// The defending player controls no Island: the attacker can't attack them.
	if canAttackTarget(g, attacker, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("attacker may attack a defender who controls no Island")
	}

	// Give the defending player an Island: now the attack is legal.
	addIslandPermanent(g, game.Player2)
	if !canAttackTarget(g, attacker, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("attacker may not attack a defender who controls an Island")
	}
}

func TestCantAttackUnlessDefenderControlsIslandChecksDefenderNotAttacker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := cantAttackUnlessDefenderControlsIsland(g, game.Player1)
	// The attacker's own controller has an Island; the defender does not.
	addIslandPermanent(g, game.Player1)
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1

	if canAttackTarget(g, attacker, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("attacker's own Island satisfied a defender-controls restriction")
	}
}

func TestCantAttackUnlessDefenderControlsIslandNilTargetIsLenient(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := cantAttackUnlessDefenderControlsIsland(g, game.Player1)
	addIslandPermanent(g, game.Player2)
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1

	// With a qualifying defender on the board, the "can it attack at all" check
	// (nil target) must not forbid the creature from attacking.
	if !canAttackWith(g, attacker, game.Player1) {
		t.Fatal("attacker cannot attack at all despite a qualifying defender")
	}
}
