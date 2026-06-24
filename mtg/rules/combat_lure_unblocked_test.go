package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func addRuleEffectSource(g *game.Game, controller game.PlayerID, kind game.RuleEffectKind, affected id.ID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Rule Effect Source",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:             kind,
				AffectedObjectID: affected,
			}},
		}},
	}})
}

func blocksOf(attacker *game.Permanent, blockers ...*game.Permanent) action.Action {
	declarations := make([]game.BlockDeclaration, 0, len(blockers))
	for _, blocker := range blockers {
		declarations = append(declarations, game.BlockDeclaration{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID})
	}
	return action.DeclareBlockers(declarations)
}

func TestTrueLureForcesEveryAbleBlocker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addRuleEffectSource(g, game.Player1, game.RuleEffectMustBeBlockedByAllAble, attacker.ObjectID)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
	}
	engine := NewEngine(nil)

	legal := legalDeclareBlockersActions(g, game.Player2)
	if len(legal) != 1 {
		t.Fatalf("legal block actions = %d, want only the all-blockers declaration", len(legal))
	}
	payload := mustDeclareBlockersPayload(t, legal[0])
	if len(payload.Blockers) != 2 {
		t.Fatalf("legal blockers = %d, want both able blockers forced onto the lure", len(payload.Blockers))
	}

	if engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, blocksOf(attacker, first))) {
		t.Fatal("applyDeclareBlockers accepted only one blocker despite true-lure requiring all able blockers")
	}
	if engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, blocksOf(attacker))) {
		t.Fatal("applyDeclareBlockers accepted no blocks despite a satisfiable true-lure requirement")
	}
	if !engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, blocksOf(attacker, first, second))) {
		t.Fatal("applyDeclareBlockers rejected the all-blockers declaration the true-lure requires")
	}
}

func TestTrueLureAllowsNoBlocksWhenUnable(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 1, game.Flying)
	addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addRuleEffectSource(g, game.Player1, game.RuleEffectMustBeBlockedByAllAble, attacker.ObjectID)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
	}
	engine := NewEngine(nil)

	legal := legalDeclareBlockersActions(g, game.Player2)
	if len(legal) != 1 {
		t.Fatalf("legal block actions = %d, want only the no-block action", len(legal))
	}
	if len(mustDeclareBlockersPayload(t, legal[0]).Blockers) != 0 {
		t.Fatal("legal blockers non-empty, want no blocks because the ground creature cannot block a flier")
	}
	if !engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, blocksOf(attacker))) {
		t.Fatal("applyDeclareBlockers rejected no blocks for an unsatisfiable true-lure requirement")
	}
}

func TestTrueLureWithMenaceUnsatisfiableFailsOpen(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	lure := addCombatCreaturePermanentWithPower(g, game.Player1, 1, game.Menace)
	other := addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addRuleEffectSource(g, game.Player1, game.RuleEffectMustBeBlockedByAllAble, lure.ObjectID)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: lure.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: other.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)

	legal := legalDeclareBlockersActions(g, game.Player2)
	sawNoBlock := false
	sawBlockOther := false
	for _, act := range legal {
		payload := mustDeclareBlockersPayload(t, act)
		if len(payload.Blockers) == 0 {
			sawNoBlock = true
		}
		if len(payload.Blockers) == 1 && payload.Blockers[0].Blocking == other.ObjectID {
			sawBlockOther = true
		}
	}
	if !sawNoBlock {
		t.Fatal("no-block action missing: an unsatisfiable menaced lure must not force blocks")
	}
	if !sawBlockOther {
		t.Fatal("blocking the non-lure attacker missing: the lure must not invalidate other legal blocks")
	}
	if !engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, blocksOf(other, blocker))) {
		t.Fatal("applyDeclareBlockers rejected a legal block of the non-lure attacker")
	}
}

func TestAssignCombatDamageAsThoughUnblockedHitsDefendingPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	blocker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Wall",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 5}),
	}})
	addRuleEffectSource(g, game.Player1, game.RuleEffectAssignCombatDamageAsThoughUnblocked, attacker.ObjectID)
	g.Combat = blockedCombat(attacker, blocker)
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.resolveCombatDamage(g, &log)

	if g.Players[game.Player2].Life != 37 {
		t.Fatalf("defending player life = %d, want 37 (blocked attacker still deals to player)", g.Players[game.Player2].Life)
	}
	if blocker.MarkedDamage != 0 {
		t.Fatalf("blocker marked damage = %d, want 0 (damage assigned to the player instead)", blocker.MarkedDamage)
	}
	if attacker.MarkedDamage != 2 {
		t.Fatalf("attacker marked damage = %d, want 2 from its blocker", attacker.MarkedDamage)
	}
}
