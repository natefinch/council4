package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
)

func TestResolveCombatWithAttackersDealsDamageWithoutMutatingOriginal(t *testing.T) {
	e := newSimEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}

	defenderLifeBefore := g.Players[game.Player2].Life
	declare := action.DeclareAttackersAction{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}

	// Pass policies mean the defender does not block, so the attacker connects.
	resolved, ok := e.Simulator().ResolveCombatWithAttackers(g, game.Player1, declare, simPassPolicies())
	if !ok {
		t.Fatal("ResolveCombatWithAttackers reported the declaration illegal")
	}
	if got := resolved.Players[game.Player2].Life; got != defenderLifeBefore-3 {
		t.Fatalf("defender life = %d, want %d (took 3 unblocked combat damage)", got, defenderLifeBefore-3)
	}
	if got := g.Players[game.Player2].Life; got != defenderLifeBefore {
		t.Fatalf("original defender life mutated: %d, want %d", got, defenderLifeBefore)
	}
	if len(g.Combat.Attackers) != 0 {
		t.Fatal("resolving combat on the clone mutated the original combat state")
	}
}

func TestResolveCombatWithAttackersTrades(t *testing.T) {
	// A 2/2 attacks; the defender's rollout policy will block with its own 2/2
	// (the generic strategy blocks to trade), so both die. Here the defender uses
	// a policy that blocks: we assert the attacker died in the trade.
	e := newSimEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}

	declare := action.DeclareAttackersAction{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	policies := simPassPolicies()
	policies[game.Player2] = blockingPolicy{blocker: blocker.ObjectID, attacker: attacker.ObjectID}

	resolved, ok := e.Simulator().ResolveCombatWithAttackers(g, game.Player1, declare, policies)
	if !ok {
		t.Fatal("ResolveCombatWithAttackers reported the declaration illegal")
	}
	if permanentStillPresent(resolved, attacker.ObjectID) {
		t.Fatal("attacker survived a lethal block; combat damage did not resolve")
	}
	if permanentStillPresent(resolved, blocker.ObjectID) {
		t.Fatal("blocker survived; the 2/2 trade did not resolve")
	}
}

func TestResolveCombatWithAttackersRejectsOutsideDeclareStep(t *testing.T) {
	e := newSimEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	if _, ok := e.Simulator().ResolveCombatWithAttackers(g, game.Player1, action.DeclareAttackersAction{}, simPassPolicies()); ok {
		t.Fatal("ResolveCombatWithAttackers accepted a call outside the declare-attackers step")
	}
}

// blockingPolicy is a defender policy that blocks a specific attacker with a
// specific creature whenever a declare-blockers action offering that assignment
// is legal, and passes otherwise.
type blockingPolicy struct {
	blocker  id.ID
	attacker id.ID
}

func (blockingPolicy) ChooseChoice(_ PlayerObservation, _ game.ChoiceRequest) []int {
	return nil
}

func (p blockingPolicy) ChooseAction(_ PlayerObservation, legal []action.Action) action.Action {
	want := game.BlockDeclaration{Blocker: p.blocker, Blocking: p.attacker}
	for _, act := range legal {
		payload, ok := act.DeclareBlockersPayload()
		if !ok {
			continue
		}
		if slices.Contains(payload.Blockers, want) {
			return act
		}
	}
	return action.Pass()
}

// permanentStillPresent reports whether a permanent with the given object ID is
// still on the battlefield.
func permanentStillPresent(g *game.Game, objectID id.ID) bool {
	_, ok := permanentByObjectID(g, objectID)
	return ok
}
