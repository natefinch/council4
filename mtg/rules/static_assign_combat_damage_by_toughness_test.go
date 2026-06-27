package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// addAssignsToughnessCreature adds a creature with distinct power and toughness
// that carries the "assigns combat damage equal to its toughness rather than its
// power" rule effect, scoped to creatures its controller controls.
func addAssignsToughnessCreature(g *game.Game, controller game.PlayerID, power, toughness int) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Toughness Brawler",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: toughness}),
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectAssignCombatDamageUsingToughness,
				AffectedController: game.ControllerYou,
				PermanentTypes:     []types.Card{types.Creature},
			}},
		}},
	}})
}

func TestAssignsCombatDamageByToughnessUnblockedUsesToughness(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addAssignsToughnessCreature(g, game.Player1, 0, 5)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if g.Players[game.Player2].Life != 35 {
		t.Fatalf("defending player life = %d, want 35 (5 toughness damage)", g.Players[game.Player2].Life)
	}
}

func TestAssignsCombatDamageByToughnessBlockedUsesToughness(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addAssignsToughnessCreature(g, game.Player1, 0, 4)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Combat = blockedCombat(attacker, blocker)

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if blocker.MarkedDamage != 4 {
		t.Fatalf("blocker marked damage = %d, want 4 (attacker toughness)", blocker.MarkedDamage)
	}
}

func TestAssignsCombatDamageByToughnessDoesNotAffectOtherCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Player1 controls the static, but Player2's attacker is unaffected and
	// still assigns damage equal to its power.
	addAssignsToughnessCreature(g, game.Player1, 0, 5)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
		},
	}

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if g.Players[game.Player1].Life != 37 {
		t.Fatalf("defending player life = %d, want 37 (3 power damage, static does not apply)", g.Players[game.Player1].Life)
	}
}
