package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestLegalDeclareBlockersActionsProductiveFirstThenNoBlocks(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	tapped := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	tapped.Tapped = true
	opponentCreature := addCombatCreaturePermanentWithPower(g, game.Player3, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}

	legal := legalDeclareBlockersActions(g, game.Player2)

	if len(legal) != 2 {
		t.Fatalf("legal declare blockers actions = %d, want 2", len(legal))
	}
	wantBlock := action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
	})
	if !actionsEqual(legal[0], wantBlock) {
		t.Fatalf("first legal block action = %+v, want %+v", legal[0], wantBlock)
	}
	noBlocks := mustDeclareBlockersPayload(t, legal[1])
	if len(noBlocks.Blockers) != 0 {
		t.Fatalf("last block action = %+v, want no blocks", noBlocks.Blockers)
	}
	for _, act := range legal {
		blockers := mustDeclareBlockersPayload(t, act)
		for _, block := range blockers.Blockers {
			if block.Blocker == tapped.ObjectID || block.Blocker == opponentCreature.ObjectID {
				t.Fatalf("ineligible blocker %v appeared in legal action %+v", block.Blocker, act)
			}
		}
	}
}

func TestCanBlockAttackerWhosePermanentTargetPhasedOut(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:    "Target Planeswalker",
		Types:   []types.Card{types.Planeswalker},
		Loyalty: opt.Val(5),
	}})
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Turn.ActivePlayer = game.Player1
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: attacker.ObjectID,
		Target: game.AttackTarget{
			Player:         game.Player2,
			PlaneswalkerID: target.ObjectID,
		},
	}}}

	if !phaseOutPermanentTree(g, target, game.Player2, make(map[game.ObjectID]bool)) {
		t.Fatal("phaseOutPermanentTree() = false")
	}
	gotTarget := g.Combat.Attackers[0].Target
	if !gotTarget.NoTarget || gotTarget.Player != game.Player2 ||
		gotTarget.PlaneswalkerID != 0 || gotTarget.BattleID != 0 {
		t.Fatalf("attack target after phasing = %+v, want no target with Player2 defending", gotTarget)
	}

	legal := legalDeclareBlockersActions(g, game.Player2)
	wantBlock := action.DeclareBlockers([]game.BlockDeclaration{{
		Blocker:  blocker.ObjectID,
		Blocking: attacker.ObjectID,
	}})
	if !slices.ContainsFunc(legal, func(candidate action.Action) bool {
		return actionsEqual(candidate, wantBlock)
	}) {
		t.Fatalf("legal block actions = %+v, want block %+v", legal, wantBlock)
	}

	startingLife := g.Players[game.Player2].Life
	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})
	if g.Players[game.Player2].Life != startingLife {
		t.Fatalf("defending player life = %d, want %d", g.Players[game.Player2].Life, startingLife)
	}
}

func TestApplyDeclareBlockersRecordsBlockerOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)
	declare := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
	}))

	if !engine.applyDeclareBlockers(g, game.Player2, declare) {
		t.Fatal("applyDeclareBlockers() = false, want true")
	}
	if !slices.Equal(g.Combat.Blockers, declare.Blockers) {
		t.Fatalf("combat blockers = %+v, want %+v", g.Combat.Blockers, declare.Blockers)
	}
	if !slices.Equal(g.Combat.BlockerOrder[attacker.ObjectID], []id.ID{blocker.ObjectID}) {
		t.Fatalf("blocker order = %+v, want [%v]", g.Combat.BlockerOrder[attacker.ObjectID], blocker.ObjectID)
	}
}

func TestApplyDeclareBlockersInvalidDoesNotMutate(t *testing.T) {
	tests := []struct {
		name    string
		declare func(*game.Game, *game.Permanent, *game.Permanent) action.DeclareBlockersAction
	}{
		{
			name: "duplicate blocker",
			declare: func(g *game.Game, attacker *game.Permanent, blocker *game.Permanent) action.DeclareBlockersAction {
				return mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
					{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
					{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID + 100},
				}))
			},
		},
		{
			name: "unknown attacker",
			declare: func(g *game.Game, attacker *game.Permanent, blocker *game.Permanent) action.DeclareBlockersAction {
				other := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
				return mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
					{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
					{Blocker: other.ObjectID, Blocking: attacker.ObjectID + 100},
				}))
			},
		},
		{
			name: "tapped blocker",
			declare: func(g *game.Game, attacker *game.Permanent, blocker *game.Permanent) action.DeclareBlockersAction {
				blocker.Tapped = true
				return mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
					{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
				}))
			},
		},
		{
			name: "attacker not attacking controller",
			declare: func(g *game.Game, attacker *game.Permanent, blocker *game.Permanent) action.DeclareBlockersAction {
				g.Combat.Attackers[0].Target.Player = game.Player3
				return mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
					{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
				}))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
			g.Turn.Phase = game.PhaseCombat
			g.Turn.Step = game.StepDeclareBlockers
			g.Combat = &game.CombatState{
				Attackers: []game.AttackDeclaration{
					{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
				},
			}
			engine := NewEngine(nil)

			if engine.applyDeclareBlockers(g, game.Player2, tt.declare(g, attacker, blocker)) {
				t.Fatal("applyDeclareBlockers() = true, want false")
			}
			if len(g.Combat.Blockers) != 0 {
				t.Fatalf("combat blockers = %+v, want none", g.Combat.Blockers)
			}
			if len(g.Combat.BlockerOrder) != 0 {
				t.Fatalf("blocker order = %+v, want empty", g.Combat.BlockerOrder)
			}
		})
	}
}

func TestApplyDeclareBlockersAllowsMultipleBlockersAndRecordsOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)
	declare := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: first.ObjectID, Blocking: attacker.ObjectID},
		{Blocker: second.ObjectID, Blocking: attacker.ObjectID},
	}))

	if !engine.applyDeclareBlockers(g, game.Player2, declare) {
		t.Fatal("applyDeclareBlockers() = false, want true")
	}
	if !slices.Equal(g.Combat.Blockers, declare.Blockers) {
		t.Fatalf("combat blockers = %+v, want %+v", g.Combat.Blockers, declare.Blockers)
	}
	wantOrder := []id.ID{first.ObjectID, second.ObjectID}
	if !slices.Equal(g.Combat.BlockerOrder[attacker.ObjectID], wantOrder) {
		t.Fatalf("blocker order = %+v, want %+v", g.Combat.BlockerOrder[attacker.ObjectID], wantOrder)
	}
}

func TestHorsemanshipBlockLegalityRequiresHorsemanship(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2, game.Horsemanship)
	ground := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	flying := addCombatCreaturePermanentWithPower(g, game.Player2, 2, game.Flying)
	reach := addCombatCreaturePermanentWithPower(g, game.Player2, 2, game.Reach)
	horse := addCombatCreaturePermanentWithPower(g, game.Player2, 2, game.Horsemanship)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)

	for _, blocker := range []*game.Permanent{ground, flying, reach} {
		block := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
			{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
		}))
		if engine.applyDeclareBlockers(g, game.Player2, block) {
			t.Fatal("a creature without horsemanship blocked a horsemanship attacker")
		}
	}
	horseBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: horse.ObjectID, Blocking: attacker.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, horseBlock) {
		t.Fatal("horsemanship blocker could not block horsemanship attacker")
	}
}

// TestHorsemanshipGrantedByContinuousEffectEnforcesBlockLegality proves that
// horsemanship granted to an attacker by a static keyword-grant continuous
// effect (the "<subject> has horsemanship" declaration) is enforced in block
// legality exactly like printed horsemanship: only blockers with horsemanship
// may block it.
func TestHorsemanshipGrantedByContinuousEffectEnforcesBlockLegality(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Horsemanship Granter",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility,
				Group: game.ObjectControlledGroup(
					game.SourcePermanentReference(),
					game.Selection{RequiredTypes: []types.Card{types.Creature}},
				),
				AddKeywords: []game.Keyword{game.Horsemanship},
			}},
		}},
	}})
	ground := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	horse := addCombatCreaturePermanentWithPower(g, game.Player2, 2, game.Horsemanship)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)

	groundBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: ground.ObjectID, Blocking: attacker.ObjectID},
	}))
	if engine.applyDeclareBlockers(g, game.Player2, groundBlock) {
		t.Fatal("a creature without horsemanship blocked an attacker granted horsemanship")
	}
	horseBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: horse.ObjectID, Blocking: attacker.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, horseBlock) {
		t.Fatal("horsemanship blocker could not block an attacker granted horsemanship")
	}
}

func TestFlyingBlockLegalityRequiresFlyingOrReach(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2, game.Flying)
	ground := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	flying := addCombatCreaturePermanentWithPower(g, game.Player2, 2, game.Flying)
	reach := addCombatCreaturePermanentWithPower(g, game.Player2, 2, game.Reach)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)

	groundBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: ground.ObjectID, Blocking: attacker.ObjectID},
	}))
	if engine.applyDeclareBlockers(g, game.Player2, groundBlock) {
		t.Fatal("ground blocker blocked flying attacker")
	}
	flyingBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: flying.ObjectID, Blocking: attacker.ObjectID},
		{Blocker: reach.ObjectID, Blocking: attacker.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, flyingBlock) {
		t.Fatal("flying and reach blockers could not block flying attacker")
	}
}

func TestLegalDeclareBlockersActionsExcludeIllegalFlyingAndSingleMenaceBlocks(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	flyingMenace := addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.Flying, game.Menace)
	ground := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	reach := addCombatCreaturePermanentWithPower(g, game.Player2, 2, game.Reach)
	flying := addCombatCreaturePermanentWithPower(g, game.Player2, 2, game.Flying)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: flyingMenace.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}

	legal := legalDeclareBlockersActions(g, game.Player2)

	if len(legal) != 2 {
		t.Fatalf("legal declare blockers actions = %d, want 2", len(legal))
	}
	wantBlock := action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: reach.ObjectID, Blocking: flyingMenace.ObjectID},
		{Blocker: flying.ObjectID, Blocking: flyingMenace.ObjectID},
	})
	if !actionsEqual(legal[0], wantBlock) {
		t.Fatalf("first legal block action = %+v, want %+v", legal[0], wantBlock)
	}
	noBlocks := mustDeclareBlockersPayload(t, legal[1])
	if len(noBlocks.Blockers) != 0 {
		t.Fatalf("last block action = %+v, want no blocks", noBlocks.Blockers)
	}
	for _, act := range legal {
		blockers := mustDeclareBlockersPayload(t, act)
		for _, block := range blockers.Blockers {
			if block.Blocker == ground.ObjectID {
				t.Fatalf("ground blocker appeared in legal flying block action %+v", act)
			}
		}
	}
}

func TestMenaceRequiresAtLeastTwoBlockers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.Menace)
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)

	singleBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: first.ObjectID, Blocking: attacker.ObjectID},
	}))
	if engine.applyDeclareBlockers(g, game.Player2, singleBlock) {
		t.Fatal("single blocker blocked menace attacker")
	}
	twoBlocks := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: first.ObjectID, Blocking: attacker.ObjectID},
		{Blocker: second.ObjectID, Blocking: attacker.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, twoBlocks) {
		t.Fatal("two blockers could not block menace attacker")
	}
}

func TestCantBeBlockedByMoreThanOneRejectsMultipleBlockers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Lone Duelist",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCantBeBlockedByMoreThanOne,
				AffectedSource: true,
			}},
		}},
	}})
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)

	legal := legalDeclareBlockersActions(g, game.Player2)
	for _, act := range legal {
		if len(mustDeclareBlockersPayload(t, act).Blockers) > 1 {
			t.Fatalf("legal block actions include a multi-blocker declaration: %+v", act)
		}
	}

	twoBlocks := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: first.ObjectID, Blocking: attacker.ObjectID},
		{Blocker: second.ObjectID, Blocking: attacker.ObjectID},
	}))
	if engine.applyDeclareBlockers(g, game.Player2, twoBlocks) {
		t.Fatal("two blockers blocked can't-be-blocked-by-more-than-one attacker")
	}
	singleBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: first.ObjectID, Blocking: attacker.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, singleBlock) {
		t.Fatal("single blocker rejected for can't-be-blocked-by-more-than-one attacker")
	}
}

// TestCantBeBlockedByMoreThanOneOnlyAffectsItsOwnAttacker confirms the rule
// effect is matched to its source permanent and does not leak onto a different
// attacker that two creatures may legally gang-block.
func TestCantBeBlockedByMoreThanOneOnlyAffectsItsOwnAttacker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	restricted := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Lone Duelist",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCantBeBlockedByMoreThanOne,
				AffectedSource: true,
			}},
		}},
	}})
	other := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: restricted.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: other.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)

	gangBlockOther := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: first.ObjectID, Blocking: other.ObjectID},
		{Blocker: second.ObjectID, Blocking: other.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, gangBlockOther) {
		t.Fatal("two blockers could not gang-block an unrestricted attacker")
	}
}

func TestMustBeBlockedRequirementRejectsNoBlocksWhenAble(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Must Block Effect",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:             game.RuleEffectMustBeBlocked,
				AffectedObjectID: attacker.ObjectID,
			}},
		}}},
	})
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
	}
	engine := NewEngine(nil)

	legal := legalDeclareBlockersActions(g, game.Player2)
	if len(legal) != 1 {
		t.Fatalf("legal block actions = %d, want only required block", len(legal))
	}
	wantBlock := action.DeclareBlockers([]game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID}})
	if !actionsEqual(legal[0], wantBlock) {
		t.Fatalf("legal action = %+v, want required block %+v", legal[0], wantBlock)
	}
	if engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, action.DeclareBlockers(nil))) {
		t.Fatal("applyDeclareBlockers accepted no blocks despite satisfiable must-block requirement")
	}
	if !engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, wantBlock)) {
		t.Fatal("applyDeclareBlockers rejected required block")
	}
}

func TestMustBeBlockedRequirementAllowsNoBlocksWhenUnable(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2, game.Flying)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Must Block Effect",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:             game.RuleEffectMustBeBlocked,
				AffectedObjectID: attacker.ObjectID,
			}},
		}}},
	})
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
	}
	engine := NewEngine(nil)

	legal := legalDeclareBlockersActions(g, game.Player2)
	if len(legal) != 1 {
		t.Fatalf("legal block actions = %d, want only no-block action", len(legal))
	}
	noBlocks := mustDeclareBlockersPayload(t, legal[0])
	if len(noBlocks.Blockers) != 0 {
		t.Fatalf("legal blockers = %+v, want no blocks because %v cannot block flying", noBlocks.Blockers, blocker.ObjectID)
	}
	if !engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, action.DeclareBlockers(nil))) {
		t.Fatal("applyDeclareBlockers rejected no blocks for unsatisfiable must-block requirement")
	}
}

func TestMustBeBlockedStaticBodyRequiresBlockWhenAble(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Provoking Bear",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 3}),
		Toughness:       opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{game.MustBeBlockedStaticBody},
	}})
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
	}
	engine := NewEngine(nil)

	legal := legalDeclareBlockersActions(g, game.Player2)
	if len(legal) != 1 {
		t.Fatalf("legal block actions = %d, want only required block", len(legal))
	}
	wantBlock := action.DeclareBlockers([]game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID}})
	if !actionsEqual(legal[0], wantBlock) {
		t.Fatalf("legal action = %+v, want required block %+v", legal[0], wantBlock)
	}
	if engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, action.DeclareBlockers(nil))) {
		t.Fatal("applyDeclareBlockers accepted no blocks despite must-be-blocked static body")
	}
	if !engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, wantBlock)) {
		t.Fatal("applyDeclareBlockers rejected required block from must-be-blocked static body")
	}
}

func TestProtectionFromColorBlockingEnforced(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers

	// Attacker has protection from red.
	pt := game.PT{Value: 2}
	attacker := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Protected Attacker",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(pt),
		Toughness:       opt.Val(pt),
		StaticAbilities: []game.StaticAbility{game.ProtectionFromColorsStaticAbility(color.Red)},
	}})
	// Red blocker — cannot block attacker.
	redBlocker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Red Creature",
		Types:     []types.Card{types.Creature},
		Colors:    []color.Color{color.Red},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})
	// Green blocker — can block.
	greenBlocker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Green Creature",
		Types:     []types.Card{types.Creature},
		Colors:    []color.Color{color.Green},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}

	if canBlockAttacker(g, redBlocker, attacker) {
		t.Fatal("red creature can block attacker with protection from red, want false")
	}
	if !canBlockAttacker(g, greenBlocker, attacker) {
		t.Fatal("green creature cannot block attacker with protection from red, want true")
	}
}

func TestProtectionFromEverythingBlockingEnforced(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers

	pt := game.PT{Value: 2}
	attacker := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Protected Attacker",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(pt),
		Toughness:       opt.Val(pt),
		StaticAbilities: []game.StaticAbility{game.ProtectionFromEverythingStaticAbility()},
	}})
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}

	if canBlockAttacker(g, blocker, attacker) {
		t.Fatal("creature can block attacker with protection from everything, want false")
	}
}

func TestLegalBlockersExcludesProtectionMatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers

	pt := game.PT{Value: 2}
	attacker := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Protected Attacker",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(pt),
		Toughness:       opt.Val(pt),
		StaticAbilities: []game.StaticAbility{game.ProtectionFromColorsStaticAbility(color.Blue)},
	}})
	blueBlocker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Blue Blocker",
		Types:     []types.Card{types.Creature},
		Colors:    []color.Color{color.Blue},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})
	redBlocker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Red Blocker",
		Types:     []types.Card{types.Creature},
		Colors:    []color.Color{color.Red},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}

	legal := legalDeclareBlockersActions(g, game.Player2)
	for _, act := range legal {
		payload, ok := act.DeclareBlockersPayload()
		if !ok {
			continue
		}
		for _, block := range payload.Blockers {
			if block.Blocker == blueBlocker.ObjectID {
				t.Fatal("blue blocker appeared in legal block actions but attacker has protection from blue")
			}
		}
	}
	// Red blocker should be allowed.
	found := false
	for _, act := range legal {
		payload, ok := act.DeclareBlockersPayload()
		if !ok {
			continue
		}
		for _, block := range payload.Blockers {
			if block.Blocker == redBlocker.ObjectID {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("red blocker not found in any legal block action, want allowed")
	}
}

// addRestrictedBlockAttacker returns a vanilla attacker carrying a single
// RuleEffectCantBeBlockedByCreaturesWith rule effect bounded by restriction.
func addRestrictedBlockAttacker(g *game.Game, controller game.PlayerID, restriction game.BlockerRestriction) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Evasive Attacker",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCantBeBlockedByCreaturesWith,
				AffectedSource:     true,
				BlockerRestriction: restriction,
			}},
		}},
	}})
}

func TestCantBeBlockedByCreaturesWithFlyingRejectsFlyingBlocker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addRestrictedBlockAttacker(g, game.Player1, game.BlockerRestriction{Kind: game.BlockerRestrictionFlying})
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

	flyingBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: flier.ObjectID, Blocking: attacker.ObjectID},
	}))
	if engine.applyDeclareBlockers(g, game.Player2, flyingBlock) {
		t.Fatal("flying blocker blocked a can't-be-blocked-by-flying attacker")
	}
	groundedBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: grounded.ObjectID, Blocking: attacker.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, groundedBlock) {
		t.Fatal("non-flying blocker rejected for can't-be-blocked-by-flying attacker")
	}
}

func TestCantBeBlockedByCreaturesWithPowerOrLessUsesThreshold(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addRestrictedBlockAttacker(g, game.Player1, game.BlockerRestriction{Kind: game.BlockerRestrictionPowerLessOrEqual, Power: 2})
	weak := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	strong := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)

	weakBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: weak.ObjectID, Blocking: attacker.ObjectID},
	}))
	if engine.applyDeclareBlockers(g, game.Player2, weakBlock) {
		t.Fatal("power-2 blocker blocked a can't-be-blocked-by-power-2-or-less attacker")
	}
	strongBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: strong.ObjectID, Blocking: attacker.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, strongBlock) {
		t.Fatal("power-3 blocker rejected for can't-be-blocked-by-power-2-or-less attacker")
	}
}

func TestCantBeBlockedByCreaturesWithPowerOrGreaterUsesThreshold(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addRestrictedBlockAttacker(g, game.Player1, game.BlockerRestriction{Kind: game.BlockerRestrictionPowerGreaterOrEqual, Power: 3})
	strong := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	weak := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)

	strongBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: strong.ObjectID, Blocking: attacker.ObjectID},
	}))
	if engine.applyDeclareBlockers(g, game.Player2, strongBlock) {
		t.Fatal("power-3 blocker blocked a can't-be-blocked-by-power-3-or-greater attacker")
	}
	weakBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: weak.ObjectID, Blocking: attacker.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, weakBlock) {
		t.Fatal("power-2 blocker rejected for can't-be-blocked-by-power-3-or-greater attacker")
	}
}

// TestCantBeBlockedByCreaturesWithOnlyAffectsItsOwnAttacker confirms the
// restriction is matched to its source and does not leak onto another attacker.
func TestCantBeBlockedByCreaturesWithOnlyAffectsItsOwnAttacker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	restricted := addRestrictedBlockAttacker(g, game.Player1, game.BlockerRestriction{Kind: game.BlockerRestrictionFlying})
	other := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	flier := addCombatCreaturePermanentWithPower(g, game.Player2, 2, game.Flying)
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
		{Blocker: flier.ObjectID, Blocking: other.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, blockOther) {
		t.Fatal("flying blocker could not block an unrestricted attacker")
	}
}

// addColoredCombatCreature returns a vanilla power-2 combat creature of the
// given single color.
func addColoredCombatCreature(g *game.Game, controller game.PlayerID, name string, c color.Color) *game.Permanent {
	pt := game.PT{Value: 2}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Colors:    []color.Color{c},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})
}

func TestCantBeBlockedByCreaturesWithColorRejectsMatchingColorBlocker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addRestrictedBlockAttacker(g, game.Player1, game.BlockerRestriction{Kind: game.BlockerRestrictionColor, Color: color.Black})
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

	blackBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: black.ObjectID, Blocking: attacker.ObjectID},
	}))
	if engine.applyDeclareBlockers(g, game.Player2, blackBlock) {
		t.Fatal("black blocker blocked a can't-be-blocked-by-black attacker")
	}
	whiteBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: white.ObjectID, Blocking: attacker.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, whiteBlock) {
		t.Fatal("white blocker rejected for can't-be-blocked-by-black attacker")
	}
}

func TestCantBeBlockedByCreaturesWithArtifactRejectsArtifactBlocker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addRestrictedBlockAttacker(g, game.Player1, game.BlockerRestriction{Kind: game.BlockerRestrictionArtifact})
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

	artifactBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: artifact.ObjectID, Blocking: attacker.ObjectID},
	}))
	if engine.applyDeclareBlockers(g, game.Player2, artifactBlock) {
		t.Fatal("artifact blocker blocked a can't-be-blocked-by-artifact attacker")
	}
	nonArtifactBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: nonArtifact.ObjectID, Blocking: attacker.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, nonArtifactBlock) {
		t.Fatal("non-artifact blocker rejected for can't-be-blocked-by-artifact attacker")
	}
}

// TestFearBlockLegalityRequiresArtifactOrBlack proves CR 702.36c: a creature
// with fear can't be blocked except by artifact creatures and/or black
// creatures. A colorless non-artifact creature cannot block; an artifact
// creature and a black creature each can.
func TestFearBlockLegalityRequiresArtifactOrBlack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers

	pt := game.PT{Value: 2}
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2, game.Fear)
	plainBlocker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Colorless Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})
	artifactBlocker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Artifact Creature",
		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})
	blackBlocker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Black Creature",
		Types:     []types.Card{types.Creature},
		Colors:    []color.Color{color.Black},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}

	if canBlockAttacker(g, plainBlocker, attacker) {
		t.Fatal("colorless non-artifact creature blocked a fear attacker, want false")
	}
	if !canBlockAttacker(g, artifactBlocker, attacker) {
		t.Fatal("artifact creature could not block a fear attacker, want true")
	}
	if !canBlockAttacker(g, blackBlocker, attacker) {
		t.Fatal("black creature could not block a fear attacker, want true")
	}
}
