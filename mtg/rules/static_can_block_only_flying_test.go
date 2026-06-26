package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// addCanBlockOnlyFlyingBlocker returns a vanilla creature carrying a single
// RuleEffectCanBlockOnlyCreaturesWith rule effect bounded by flying, modeling
// "This creature can block only creatures with flying." (Cloud Sprite).
func addCanBlockOnlyFlyingBlocker(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Selective Blocker",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: game.SimpleKeywords(game.Flying),
		}, {
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCanBlockOnlyCreaturesWith,
				AffectedSource:     true,
				BlockerRestriction: game.BlockerRestriction{Kind: game.BlockerRestrictionFlying},
			}},
		}},
	}})
}

// TestCanBlockOnlyCreaturesWithFlyingRejectsGroundedAttacker confirms a creature
// that "can block only creatures with flying" may block a flying attacker but
// not a grounded one.
func TestCanBlockOnlyCreaturesWithFlyingRejectsGroundedAttacker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	blocker := addCanBlockOnlyFlyingBlocker(g, game.Player2)
	flyer := addCombatCreaturePermanentWithPower(g, game.Player1, 2, game.Flying)
	grounded := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: flyer.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: grounded.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)

	groundedBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: blocker.ObjectID, Blocking: grounded.ObjectID},
	}))
	if engine.applyDeclareBlockers(g, game.Player2, groundedBlock) {
		t.Fatal("can-block-only-flying creature blocked a grounded attacker")
	}
	flyingBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: blocker.ObjectID, Blocking: flyer.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, flyingBlock) {
		t.Fatal("can-block-only-flying creature could not block a flying attacker")
	}
}

// TestCanBlockOnlyCreaturesWithFlyingOnlyAffectsItsOwnController confirms the
// permission restriction is matched to its source creature and does not leak
// onto another blocker.
func TestCanBlockOnlyCreaturesWithFlyingOnlyAffectsItsSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	restricted := addCanBlockOnlyFlyingBlocker(g, game.Player2)
	unrestricted := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	grounded := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: grounded.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)

	block := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: unrestricted.ObjectID, Blocking: grounded.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, block) {
		t.Fatal("unrestricted blocker could not block a grounded attacker")
	}
	if restricted == nil {
		t.Fatal("restricted blocker missing")
	}
}
