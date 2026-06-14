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
