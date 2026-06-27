package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// addExceptByAttacker returns a vanilla attacker carrying a single
// RuleEffectCantBeBlockedExceptBy rule effect bounded by restriction: only
// blockers matching restriction may block it; every other blocker is prohibited.
func addExceptByAttacker(g *game.Game, controller game.PlayerID, restriction game.BlockerRestriction) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Evasive Attacker",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCantBeBlockedExceptBy,
				AffectedSource:     true,
				BlockerRestriction: restriction,
			}},
		}},
	}})
}

func TestCantBeBlockedExceptByFlyingAllowsOnlyFlyingBlocker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addExceptByAttacker(g, game.Player1, game.BlockerRestriction{Kind: game.BlockerRestrictionFlying})
	flier := addCombatCreaturePermanentWithPower(g, game.Player2, 2, game.Flying)
	grounded := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)

	groundedBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: grounded.ObjectID, Blocking: attacker.ObjectID},
	}))
	if engine.applyDeclareBlockers(g, game.Player2, groundedBlock) {
		t.Fatal("non-flying blocker blocked a can't-be-blocked-except-by-flying attacker")
	}
	flyingBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: flier.ObjectID, Blocking: attacker.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, flyingBlock) {
		t.Fatal("flying blocker rejected for can't-be-blocked-except-by-flying attacker")
	}
}

func TestCantBeBlockedExceptByColorAllowsOnlyMatchingColorBlocker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addExceptByAttacker(g, game.Player1, game.BlockerRestriction{Kind: game.BlockerRestrictionColor, Color: color.Black})
	black := addColoredCombatCreature(g, game.Player2, "Black Creature", color.Black)
	white := addColoredCombatCreature(g, game.Player2, "White Creature", color.White)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)

	whiteBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: white.ObjectID, Blocking: attacker.ObjectID},
	}))
	if engine.applyDeclareBlockers(g, game.Player2, whiteBlock) {
		t.Fatal("white blocker blocked a can't-be-blocked-except-by-black attacker")
	}
	blackBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: black.ObjectID, Blocking: attacker.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, blackBlock) {
		t.Fatal("black blocker rejected for can't-be-blocked-except-by-black attacker")
	}
}

func TestCantBeBlockedExceptByArtifactAllowsOnlyArtifactBlocker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addExceptByAttacker(g, game.Player1, game.BlockerRestriction{Kind: game.BlockerRestrictionArtifact})
	pt := game.PT{Value: 2}
	artifact := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Artifact Creature",
		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})
	nonArtifact := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)

	nonArtifactBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: nonArtifact.ObjectID, Blocking: attacker.ObjectID},
	}))
	if engine.applyDeclareBlockers(g, game.Player2, nonArtifactBlock) {
		t.Fatal("non-artifact blocker blocked a can't-be-blocked-except-by-artifact attacker")
	}
	artifactBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: artifact.ObjectID, Blocking: attacker.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, artifactBlock) {
		t.Fatal("artifact blocker rejected for can't-be-blocked-except-by-artifact attacker")
	}
}

func TestCantBeBlockedExceptByDefenderAllowsOnlyDefenderBlocker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addExceptByAttacker(g, game.Player1, game.BlockerRestriction{Kind: game.BlockerRestrictionDefender})
	wall := addCombatCreaturePermanentWithPower(g, game.Player2, 2, game.Defender)
	plain := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)

	plainBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: plain.ObjectID, Blocking: attacker.ObjectID},
	}))
	if engine.applyDeclareBlockers(g, game.Player2, plainBlock) {
		t.Fatal("non-defender blocker blocked a can't-be-blocked-except-by-defender attacker")
	}
	wallBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: wall.ObjectID, Blocking: attacker.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, wallBlock) {
		t.Fatal("defender blocker rejected for can't-be-blocked-except-by-defender attacker")
	}
}

func TestCantBeBlockedExceptByLegendaryAllowsOnlyLegendaryBlocker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addExceptByAttacker(g, game.Player1, game.BlockerRestriction{Kind: game.BlockerRestrictionLegendary})
	pt := game.PT{Value: 2}
	legendary := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:       "Legendary Creature",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Power:      opt.Val(pt),
		Toughness:  opt.Val(pt),
	}})
	nonLegendary := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)

	nonLegendaryBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: nonLegendary.ObjectID, Blocking: attacker.ObjectID},
	}))
	if engine.applyDeclareBlockers(g, game.Player2, nonLegendaryBlock) {
		t.Fatal("non-legendary blocker blocked a can't-be-blocked-except-by-legendary attacker")
	}
	legendaryBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: legendary.ObjectID, Blocking: attacker.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, legendaryBlock) {
		t.Fatal("legendary blocker rejected for can't-be-blocked-except-by-legendary attacker")
	}
}

// TestCantBeBlockedExceptByOnlyAffectsItsOwnAttacker confirms the restriction is
// matched to its source and does not leak onto another, unrestricted attacker.
func TestCantBeBlockedExceptByOnlyAffectsItsOwnAttacker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	restricted := addExceptByAttacker(g, game.Player1, game.BlockerRestriction{Kind: game.BlockerRestrictionFlying})
	other := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	grounded := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: restricted.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: other.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)

	blockOther := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: grounded.ObjectID, Blocking: other.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, blockOther) {
		t.Fatal("grounded blocker could not block an unrestricted attacker")
	}
}
