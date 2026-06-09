package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestCombatPhaseVisitsPriorityStepsInOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	recorder := &combatStepRecorder{}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: recorder,
		game.Player2: recorder,
		game.Player3: recorder,
		game.Player4: recorder,
	}

	engine.runCombatPhase(g, agents, &TurnLog{})

	want := []game.Step{
		game.StepBeginningOfCombat,
		game.StepDeclareAttackers,
		game.StepDeclareBlockers,
		game.StepCombatDamage,
		game.StepEndOfCombat,
	}
	if !slices.Equal(recorder.firstVisits, want) {
		t.Fatalf("visited combat steps = %v, want %v", recorder.firstVisits, want)
	}
}

func TestCombatPhaseInitializesAndClearsCombatState(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	recorder := &combatStateRecorder{game: g}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: recorder,
		game.Player2: recorder,
		game.Player3: recorder,
		game.Player4: recorder,
	}

	engine.runCombatPhase(g, agents, &TurnLog{})

	if !recorder.sawCombatState {
		t.Fatal("agent never observed initialized combat state during combat")
	}
	if g.Combat != nil {
		t.Fatalf("combat state after combat = %+v, want nil", g.Combat)
	}
}

func TestCombatPhasePriorityWindowsPassThroughWithoutActions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.runCombatPhase(g, allFirstLegalAgents(), &log)

	if g.Turn.Phase != game.PhaseCombat {
		t.Fatalf("phase = %v, want %v", g.Turn.Phase, game.PhaseCombat)
	}
	if g.Turn.Step != game.StepEndOfCombat {
		t.Fatalf("step = %v, want %v", g.Turn.Step, game.StepEndOfCombat)
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0", g.Stack.Size())
	}
	passCount := 0
	declareAttackersCount := 0
	for _, logged := range log.Actions {
		switch logged.Action.Kind {
		case action.ActionPass:
			passCount++
		case action.ActionDeclareAttackers:
			declareAttackersCount++
		default:
			t.Fatalf("logged action kind = %v, want pass or declare attackers", logged.Action.Kind)
		}
	}
	if passCount != game.NumPlayers*5 {
		t.Fatalf("logged pass actions = %d, want %d", passCount, game.NumPlayers*5)
	}
	if declareAttackersCount != 1 {
		t.Fatalf("logged declare attackers actions = %d, want 1", declareAttackersCount)
	}
}

func TestEligibleAttackersFiltersIllegalCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	eligible := addCombatCreaturePermanent(g, game.Player1)
	tapped := addCombatCreaturePermanent(g, game.Player1)
	tapped.Tapped = true
	sick := addCombatCreaturePermanent(g, game.Player1)
	sick.SummoningSick = true
	hasty := addCombatCreaturePermanent(g, game.Player1, game.Haste)
	hasty.SummoningSick = true
	defender := addCombatCreaturePermanent(g, game.Player1, game.Defender)
	opponent := addCombatCreaturePermanent(g, game.Player2)
	nonCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Relic",
		Types: []types.Card{types.Artifact}},
	})

	got := eligibleAttackers(g, game.Player1)

	if !slices.Equal(got, []*game.Permanent{eligible, hasty}) {
		t.Fatalf("eligible attackers = %v, want [%v %v]", permanentIDs(got), eligible.ObjectID, hasty.ObjectID)
	}
	for _, permanent := range []*game.Permanent{tapped, sick, defender, opponent, nonCreature} {
		if slices.Contains(got, permanent) {
			t.Fatalf("ineligible permanent %v was eligible", permanent.ObjectID)
		}
	}
}

func TestLegalDeclareAttackersActionsProductiveFirstThenNoAttacks(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker1 := addCombatCreaturePermanent(g, game.Player1)
	attacker2 := addCombatCreaturePermanent(g, game.Player1)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	g.Players[game.Player3].Eliminated = true
	g.TurnOrder.Eliminate(game.Player3)

	legal := legalDeclareAttackersActions(g, game.Player1)

	if len(legal) != 7 {
		t.Fatalf("legal declare attackers actions = %d, want 7", len(legal))
	}
	wantTargets := []game.PlayerID{game.Player2, game.Player4}
	for targetIndex, target := range wantTargets {
		allAttackersAction := targetIndex*3 + 2
		for i := targetIndex * 3; i <= allAttackersAction; i++ {
			if legal[i].Kind != action.ActionDeclareAttackers {
				t.Fatalf("action %d kind = %v, want declare attackers", i, legal[i].Kind)
			}
		}
		want := []game.AttackDeclaration{
			{Attacker: attacker1.ObjectID, Target: game.AttackTarget{Player: target}},
			{Attacker: attacker2.ObjectID, Target: game.AttackTarget{Player: target}},
		}
		attackers := mustDeclareAttackersPayload(t, legal[allAttackersAction])
		if !slices.Equal(attackers.Attackers, want) {
			t.Fatalf("action %d attackers = %+v, want %+v", allAttackersAction, attackers.Attackers, want)
		}
	}
	attackers := mustDeclareAttackersPayload(t, legal[6])
	if len(attackers.Attackers) != 0 {
		t.Fatalf("last declare attackers action = %+v, want no attacks", attackers.Attackers)
	}
}

func TestGoadedCreatureMustAttackIfAble(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanent(g, game.Player1)
	attacker.Goaded = map[game.PlayerID]game.GoadStatus{game.Player2: {CreatedTurn: 1, ExpiresFor: game.Player2}}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	engine := NewEngine(nil)

	legal := legalDeclareAttackersActions(g, game.Player1)

	for _, act := range legal {
		attackers := mustDeclareAttackersPayload(t, act)
		if len(attackers.Attackers) == 0 {
			t.Fatalf("legal actions included no attacks despite goaded eligible attacker: %+v", legal)
		}
	}
	if engine.applyDeclareAttackers(g, game.Player1, mustDeclareAttackersPayload(t, action.DeclareAttackers(nil))) {
		t.Fatal("applyDeclareAttackers() accepted no attacks with goaded eligible attacker")
	}
	legalAttack := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player3}},
	}))
	if !engine.applyDeclareAttackers(g, game.Player1, legalAttack) {
		t.Fatal("applyDeclareAttackers() rejected legal goaded attack")
	}
}

func TestGoadedCreatureAttacksNonGoadingPlayerIfAble(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanent(g, game.Player1)
	attacker.Goaded = map[game.PlayerID]game.GoadStatus{game.Player2: {CreatedTurn: 1, ExpiresFor: game.Player2}}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	engine := NewEngine(nil)

	legal := legalDeclareAttackersActions(g, game.Player1)

	for _, act := range legal {
		attackers := mustDeclareAttackersPayload(t, act)
		for _, attack := range attackers.Attackers {
			if attack.Target.Player == game.Player2 {
				t.Fatalf("legal actions included attack at goading player while alternatives exist: %+v", legal)
			}
		}
	}
	goadingAttack := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}))
	if engine.applyDeclareAttackers(g, game.Player1, goadingAttack) {
		t.Fatal("applyDeclareAttackers() accepted goaded attack at goading player while alternatives exist")
	}
	nonGoadingAttack := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player4}},
	}))
	if !engine.applyDeclareAttackers(g, game.Player1, nonGoadingAttack) {
		t.Fatal("applyDeclareAttackers() rejected attack at non-goading player")
	}
}

func TestGoadedByTwoPlayersMustAttackRemainingNonGoadingOpponentIfAble(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanent(g, game.Player1)
	attacker.Goaded = map[game.PlayerID]game.GoadStatus{
		game.Player2: {CreatedTurn: 1, ExpiresFor: game.Player2},
		game.Player3: {CreatedTurn: 1, ExpiresFor: game.Player3},
	}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	engine := NewEngine(nil)

	legal := legalDeclareAttackersActions(g, game.Player1)

	if len(legal) != 1 {
		t.Fatalf("legal actions = %d, want only attack at remaining non-goading opponent", len(legal))
	}
	want := action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player4}},
	})
	if !actionsEqual(legal[0], want) {
		t.Fatalf("legal action = %+v, want %+v", legal[0], want)
	}
	goadingAttack := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}))
	if engine.applyDeclareAttackers(g, game.Player1, goadingAttack) {
		t.Fatal("applyDeclareAttackers() accepted attack at goading player while remaining opponent exists")
	}
	if !engine.applyDeclareAttackers(g, game.Player1, mustDeclareAttackersPayload(t, want)) {
		t.Fatal("applyDeclareAttackers() rejected attack at remaining non-goading opponent")
	}
}

func TestGoadDoesNotForceIllegalAttacks(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	defender := addCombatCreaturePermanent(g, game.Player1, game.Defender)
	defender.Goaded = map[game.PlayerID]game.GoadStatus{game.Player2: {CreatedTurn: 1, ExpiresFor: game.Player2}}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	engine := NewEngine(nil)

	legal := legalDeclareAttackersActions(g, game.Player1)

	attackers := mustDeclareAttackersPayload(t, legal[0])
	if len(legal) != 1 || len(attackers.Attackers) != 0 {
		t.Fatalf("legal actions = %+v, want only no attacks", legal)
	}
	if !engine.applyDeclareAttackers(g, game.Player1, mustDeclareAttackersPayload(t, action.DeclareAttackers(nil))) {
		t.Fatal("applyDeclareAttackers() rejected no attacks when goaded creature could not legally attack")
	}
}

func TestApplyDeclareAttackersTapsNormalButNotVigilanceAttackers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	normal := addCombatCreaturePermanent(g, game.Player1)
	vigilance := addCombatCreaturePermanent(g, game.Player1, game.Vigilance)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	engine := NewEngine(nil)
	declare := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: normal.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: vigilance.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}))

	if !engine.applyDeclareAttackers(g, game.Player1, declare) {
		t.Fatal("applyDeclareAttackers() = false, want true")
	}
	if !normal.Tapped {
		t.Fatal("normal attacker was not tapped")
	}
	if vigilance.Tapped {
		t.Fatal("vigilance attacker was tapped")
	}
	if !slices.Equal(g.Combat.Attackers, declare.Attackers) {
		t.Fatalf("combat attackers = %+v, want %+v", g.Combat.Attackers, declare.Attackers)
	}
}

func TestApplyDeclareAttackersInvalidDoesNotMutate(t *testing.T) {
	tests := []struct {
		name    string
		declare func(*game.Game, *game.Permanent) action.DeclareAttackersAction
	}{
		{
			name: "duplicate attacker",
			declare: func(g *game.Game, attacker *game.Permanent) action.DeclareAttackersAction {
				return mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
					{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
					{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player3}},
				}))
			},
		},
		{
			name: "dead defending player",
			declare: func(g *game.Game, attacker *game.Permanent) action.DeclareAttackersAction {
				g.Players[game.Player2].Eliminated = true
				g.TurnOrder.Eliminate(game.Player2)
				return mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
					{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
				}))
			},
		},
		{
			name: "planeswalker target",
			declare: func(g *game.Game, attacker *game.Permanent) action.DeclareAttackersAction {
				return mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
					{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2, PlaneswalkerID: 99}},
				}))
			},
		},
		{
			name: "summoning sick attacker",
			declare: func(g *game.Game, attacker *game.Permanent) action.DeclareAttackersAction {
				attacker.SummoningSick = true
				return mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
					{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
				}))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			attacker := addCombatCreaturePermanent(g, game.Player1)
			g.Turn.Phase = game.PhaseCombat
			g.Turn.Step = game.StepDeclareAttackers
			g.Combat = &game.CombatState{}
			engine := NewEngine(nil)

			if engine.applyDeclareAttackers(g, game.Player1, tt.declare(g, attacker)) {
				t.Fatal("applyDeclareAttackers() = true, want false")
			}
			if len(g.Combat.Attackers) != 0 {
				t.Fatalf("combat attackers = %+v, want none", g.Combat.Attackers)
			}
			if attacker.Tapped {
				t.Fatal("attacker was tapped by invalid declaration")
			}
		})
	}
}

func TestDeclareAttackersCanTargetPlaneswalkersAndBattles(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanent(g, game.Player1)
	planeswalker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Test Planeswalker",
		Types:   []types.Card{types.Planeswalker},
		Loyalty: opt.Val(3)},
	})
	battle := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Name: "Test Battle",
		Types:   []types.Card{types.Battle},
		Defense: opt.Val(4)},
	})
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	engine := NewEngine(nil)

	legal := legalDeclareAttackersActions(g, game.Player1)

	wantPlaneswalker := game.AttackTarget{Player: game.Player2, PlaneswalkerID: planeswalker.ObjectID}
	wantBattle := game.AttackTarget{Player: game.Player3, BattleID: battle.ObjectID}
	if !declareAttackersActionsContainTarget(legal, attacker.ObjectID, wantPlaneswalker) {
		t.Fatalf("legal actions = %+v, want planeswalker target %v", legal, wantPlaneswalker)
	}
	if !declareAttackersActionsContainTarget(legal, attacker.ObjectID, wantBattle) {
		t.Fatalf("legal actions = %+v, want battle target %v", legal, wantBattle)
	}
	declare := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: wantPlaneswalker},
	}))
	if !engine.applyDeclareAttackers(g, game.Player1, declare) {
		t.Fatal("applyDeclareAttackers() rejected valid planeswalker target")
	}
}

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

func TestResolveCombatDamageReducesDefendingPlayerLife(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.resolveCombatDamage(g, &log)

	if g.Players[game.Player2].Life != 38 {
		t.Fatalf("defending player life = %d, want 38", g.Players[game.Player2].Life)
	}
	if len(log.CombatDamage) != 1 {
		t.Fatalf("combat damage logs = %d, want 1", len(log.CombatDamage))
	}
	got := log.CombatDamage[0]
	if got.Attacker != attacker.ObjectID ||
		got.SourceID != attacker.CardInstanceID ||
		got.Controller != game.Player1 ||
		got.DefendingPlayer != game.Player2 ||
		got.Damage != 2 {
		t.Fatalf("combat damage log = %+v, want attacker %v source %v controller %v defender %v damage 2",
			got, attacker.ObjectID, attacker.CardInstanceID, game.Player1, game.Player2)
	}
}

func TestResolveCombatDamageMultipleAttackersDealSeparateDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker1 := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	attacker2 := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker1.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: attacker2.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.resolveCombatDamage(g, &log)

	if g.Players[game.Player2].Life != 35 {
		t.Fatalf("defending player life = %d, want 35", g.Players[game.Player2].Life)
	}
	if len(log.CombatDamage) != 2 {
		t.Fatalf("combat damage logs = %d, want 2", len(log.CombatDamage))
	}
	if log.CombatDamage[0].Damage != 2 || log.CombatDamage[1].Damage != 3 {
		t.Fatalf("combat damage log amounts = [%d %d], want [2 3]", log.CombatDamage[0].Damage, log.CombatDamage[1].Damage)
	}
}

func TestAttackerChosenCombatDamageAssignmentIsUsed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		Blockers: []game.BlockDeclaration{
			{Blocker: first.ObjectID, Blocking: attacker.ObjectID},
			{Blocker: second.ObjectID, Blocking: attacker.ObjectID},
		},
		DamageAssignment: map[id.ID]int{
			first.ObjectID:  4,
			second.ObjectID: 1,
		},
	}

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if first.MarkedDamage != 4 || second.MarkedDamage != 1 {
		t.Fatalf("blocker damage = %d/%d, want attacker-chosen 4/1", first.MarkedDamage, second.MarkedDamage)
	}
}

func TestOutOfOrderCombatDamageAssignmentFallsBackToDeterministicAssignment(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		Blockers: []game.BlockDeclaration{
			{Blocker: first.ObjectID, Blocking: attacker.ObjectID},
			{Blocker: second.ObjectID, Blocking: attacker.ObjectID},
		},
		DamageAssignment: map[id.ID]int{
			first.ObjectID:  1,
			second.ObjectID: 4,
		},
	}

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if first.MarkedDamage != 3 || second.MarkedDamage != 2 {
		t.Fatalf("blocker damage = %d/%d, want deterministic fallback 3/2", first.MarkedDamage, second.MarkedDamage)
	}
}

func TestUnderAssignedCombatDamageFallsBackToDeterministicAssignment(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		Blockers: []game.BlockDeclaration{
			{Blocker: first.ObjectID, Blocking: attacker.ObjectID},
			{Blocker: second.ObjectID, Blocking: attacker.ObjectID},
		},
		DamageAssignment: map[id.ID]int{
			first.ObjectID: 1,
		},
	}

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if first.MarkedDamage != 3 || second.MarkedDamage != 2 {
		t.Fatalf("blocker damage = %d/%d, want deterministic fallback 3/2", first.MarkedDamage, second.MarkedDamage)
	}
}

func TestAttackerChosenTrampleDeathtouchAssignmentCarriesExcessToPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 5, game.Trample, game.Deathtouch)
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		Blockers: []game.BlockDeclaration{
			{Blocker: first.ObjectID, Blocking: attacker.ObjectID},
			{Blocker: second.ObjectID, Blocking: attacker.ObjectID},
		},
		DamageAssignment: map[id.ID]int{
			first.ObjectID:  1,
			second.ObjectID: 1,
		},
	}

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if first.MarkedDamage != 1 || second.MarkedDamage != 1 {
		t.Fatalf("blocker damage = %d/%d, want deathtouch lethal 1/1", first.MarkedDamage, second.MarkedDamage)
	}
	if g.Players[game.Player2].Life != 37 {
		t.Fatalf("defending player life = %d, want 3 trample damage", g.Players[game.Player2].Life)
	}
}

func TestLegalDeclareAttackersIncludesSingleAttackerChoices(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1

	actions := legalDeclareAttackersActions(g, game.Player1)

	if !declareAttackersActionsContainTarget(actions, first.ObjectID, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("legal attacks did not include first creature attacking alone")
	}
	if !declareAttackersActionsContainTarget(actions, second.ObjectID, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("legal attacks did not include second creature attacking alone")
	}
}

func TestPhasedOutCreatureCannotAttackBlockOrBeAttacked(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	planeswalker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Phased Walker",
		Types:   []types.Card{types.Planeswalker},
		Loyalty: opt.Val(3)},
	})
	attacker.PhasedOut = true
	blocker.PhasedOut = true
	planeswalker.PhasedOut = true
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1

	if canAttackWith(g, attacker, game.Player1) {
		t.Fatal("phased-out creature can attack")
	}
	if canBlockWith(g, blocker, game.Player2) {
		t.Fatal("phased-out creature can block")
	}
	for _, target := range legalAttackTargets(g, game.Player1) {
		if target.PlaneswalkerID == planeswalker.ObjectID {
			t.Fatal("phased-out planeswalker is an attack target")
		}
	}
}

func TestStaticRuleEffectsCanProhibitAttackingAndBlocking(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Pacifying Law",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{
				{
					Kind:               game.RuleEffectCantAttack,
					AffectedController: game.ControllerOpponent,
					PermanentTypes:     []types.Card{types.Creature},
				},
				{
					Kind:               game.RuleEffectCantBlock,
					AffectedController: game.ControllerOpponent,
					PermanentTypes:     []types.Card{types.Creature},
				},
			},
		}}},
	})
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player2

	if canAttackWith(g, attacker, game.Player2) {
		t.Fatal("opponent creature could attack through cant-attack rule effect")
	}
	if canBlockWith(g, blocker, game.Player2) {
		t.Fatal("opponent creature could block through cant-block rule effect")
	}
}

func TestCantBlockStaticBodyProhibitsSourceFromBlocking(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	blocker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:            "Reluctant Bear",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 3}),
		Toughness:       opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{game.CantBlockStaticBody},
	}})
	otherBlocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	if canBlockWith(g, blocker, game.Player2) {
		t.Fatal("source with cannot-block static ability could block")
	}
	if !canBlockWith(g, otherBlocker, game.Player2) {
		t.Fatal("cannot-block static ability affected another creature")
	}
}

func TestCantBeBlockedStaticBodyProhibitsBlockingSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Elusive Bear",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 2}),
		Toughness:       opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{game.CantBeBlockedStaticBody},
	}})
	otherAttacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	if canBlockAttacker(g, blocker, attacker) {
		t.Fatal("source with cannot-be-blocked static ability could be blocked")
	}
	if !canBlockAttacker(g, blocker, otherAttacker) {
		t.Fatal("cannot-be-blocked static ability affected another creature")
	}
}

func TestCantAttackRuleCanApplyOnlyToSpecificDefender(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "No Attacks Here",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCantAttack,
				AffectedController: game.ControllerOpponent,
				PermanentTypes:     []types.Card{types.Creature},
				DefendingPlayer:    game.PlayerYou,
			}},
		}}},
	})

	if !canAttackWith(g, attacker, game.Player2) {
		t.Fatal("target-specific cant-attack effect should not remove attack eligibility")
	}
	if canAttackTarget(g, attacker, game.AttackTarget{Player: game.Player1}) {
		t.Fatal("creature could attack protected player")
	}
	if !canAttackTarget(g, attacker, game.AttackTarget{Player: game.Player3}) {
		t.Fatal("creature could not attack unprotected player")
	}
}

func TestEliminatedPlayerCleanupRemovesCombatAndStackObjects(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	owned := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	controlled := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	controlled.Controller = game.Player2
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		Blockers:  []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID}},
	}
	g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), Controller: game.Player2})

	engine.eliminatePlayer(g, game.Player2)

	if len(g.Combat.Attackers) != 0 || len(g.Combat.Blockers) != 0 {
		t.Fatalf("combat after elimination attackers=%+v blockers=%+v, want cleared", g.Combat.Attackers, g.Combat.Blockers)
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size after elimination = %d, want 0", g.Stack.Size())
	}
	if _, ok := permanentByObjectID(g, owned.ObjectID); ok || !g.Players[game.Player2].Exile.Contains(owned.CardInstanceID) {
		t.Fatal("eliminated player's owned permanent did not leave battlefield")
	}
	if _, ok := permanentByObjectID(g, controlled.ObjectID); !ok || controlled.Controller != game.Player1 {
		t.Fatalf("controlled permanent after elimination = %+v, want returned to owner control", controlled)
	}
}

func TestAttackTaxFiltersAndChargesDeclareAttackers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.AttackTaxes = append(g.AttackTaxes, game.AttackTax{DefendingPlayer: game.Player2, Amount: 1})
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1

	actions := legalDeclareAttackersActions(g, game.Player1)
	if declareAttackersActionsContainTarget(actions, attacker.ObjectID, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("taxed attack was legal without mana")
	}
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	actions = legalDeclareAttackersActions(g, game.Player1)
	if !declareAttackersActionsContainTarget(actions, attacker.ObjectID, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("taxed attack was not legal with mana available")
	}

	declare := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}}))
	if !engine.applyDeclareAttackers(g, game.Player1, declare) {
		t.Fatal("applyDeclareAttackers() = false, want tax payment to succeed")
	}
	if !forest.Tapped {
		t.Fatal("attack tax did not tap mana source")
	}
}

func TestAttackTaxCannotBePaidByDeclaredAttackerManaAbility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	manaDork := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Mana Dork",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 1}),
		Toughness:       opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{game.HasteStaticBody}},
	}, mana.G, 1)
	manaDork.SummoningSick = false
	g.AttackTaxes = append(g.AttackTaxes, game.AttackTax{DefendingPlayer: game.Player2, Amount: 1})
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1

	actions := legalDeclareAttackersActions(g, game.Player1)

	if declareAttackersActionsContainTarget(actions, manaDork.ObjectID, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("taxed attack was legal by using the declared attacker as its own mana source")
	}
}

func TestLifelinkGainsLifeFromCombatDamageToPlayers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.Lifelink)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.resolveCombatDamage(g, &log)

	if g.Players[game.Player1].Life != 43 {
		t.Fatalf("attacking player life = %d, want 43", g.Players[game.Player1].Life)
	}
	if g.Players[game.Player2].Life != 37 {
		t.Fatalf("defending player life = %d, want 37", g.Players[game.Player2].Life)
	}
}

func TestLifelinkGainsLifeFromCombatDamageToCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.Lifelink)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
		},
	}
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.resolveCombatDamage(g, &log)

	if g.Players[game.Player1].Life != 43 {
		t.Fatalf("attacking player life = %d, want 43", g.Players[game.Player1].Life)
	}
	if blocker.MarkedDamage != 3 {
		t.Fatalf("blocker marked damage = %d, want 3", blocker.MarkedDamage)
	}
	if attacker.MarkedDamage != 2 {
		t.Fatalf("attacker marked damage = %d, want 2", attacker.MarkedDamage)
	}
}

func TestCombatDamageToPlaneswalkerRemovesLoyaltyAndSBA(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	planeswalker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Test Planeswalker",
		Types:   []types.Card{types.Planeswalker},
		Loyalty: opt.Val(3)},
	})
	planeswalker.Counters.Add(counter.Loyalty, 3)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2, PlaneswalkerID: planeswalker.ObjectID}},
		},
	}
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.resolveCombatDamage(g, &log)
	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if planeswalker.Counters.Get(counter.Loyalty) != 0 {
		t.Fatalf("planeswalker loyalty = %d, want 0", planeswalker.Counters.Get(counter.Loyalty))
	}
	if _, ok := permanentByObjectID(g, planeswalker.ObjectID); ok {
		t.Fatal("zero-loyalty planeswalker remained on battlefield")
	}
	if len(deaths) != 1 || deaths[0].Reason != PermanentDeathReasonZeroLoyalty {
		t.Fatalf("deaths = %+v, want one zero-loyalty death", deaths)
	}
	if g.Players[game.Player2].Life != 40 {
		t.Fatalf("defending player life = %d, want unchanged 40", g.Players[game.Player2].Life)
	}
	if len(log.CreatureDamage) != 1 || log.CreatureDamage[0].DamagedPermanent != planeswalker.ObjectID {
		t.Fatalf("creature damage logs = %+v, want planeswalker damage", log.CreatureDamage)
	}
}

func TestCombatDamageToBattleRemovesDefenseAndSBA(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 4)
	battle := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Test Battle",
		Types:   []types.Card{types.Battle},
		Defense: opt.Val(4)},
	})
	battle.Counters.Add(counter.Defense, 4)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2, BattleID: battle.ObjectID}},
		},
	}
	engine := NewEngine(nil)

	engine.resolveCombatDamage(g, &TurnLog{})
	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if battle.Counters.Get(counter.Defense) != 0 {
		t.Fatalf("battle defense = %d, want 0", battle.Counters.Get(counter.Defense))
	}
	if _, ok := permanentByObjectID(g, battle.ObjectID); ok {
		t.Fatal("zero-defense battle remained on battlefield")
	}
	if len(deaths) != 1 || deaths[0].Reason != PermanentDeathReasonZeroDefense {
		t.Fatalf("deaths = %+v, want one zero-defense death", deaths)
	}
}

func TestCommanderCombatDamageEliminatesPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	commander := addCombatCreaturePermanentWithPower(g, game.Player1, 21)
	g.Players[game.Player1].CommanderInstanceID = commander.CardInstanceID
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.runCombatPhase(g, allFirstLegalAgents(), &log)

	if got := g.Players[game.Player2].CommanderDamage[commander.CardInstanceID]; got != 21 {
		t.Fatalf("commander damage = %d, want 21", got)
	}
	if !g.Players[game.Player2].Eliminated {
		t.Fatal("defending player was not eliminated by commander damage")
	}
	if len(log.Losses) != 1 || log.Losses[0].Player != game.Player2 || log.Losses[0].Reason != LossReasonCommanderDamage {
		t.Fatalf("loss logs = %+v, want Player2 commander damage loss", log.Losses)
	}
}

func TestNonCommanderCombatDamageDoesNotTrackCommanderDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	commander := addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
	g.Players[game.Player1].CommanderInstanceID = commander.CardInstanceID
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: creature.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}

	engine := NewEngine(nil)
	log := TurnLog{}

	engine.resolveCombatDamage(g, &log)

	if len(g.Players[game.Player2].CommanderDamage) != 0 {
		t.Fatalf("commander damage = %+v, want none", g.Players[game.Player2].CommanderDamage)
	}
}

func TestStolenCommanderStillDealsCommanderDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	commander := addCombatCreaturePermanentWithPower(g, game.Player1, 7)
	g.Players[game.Player1].CommanderInstanceID = commander.CardInstanceID
	g.CommanderIDs = map[id.ID]bool{commander.CardInstanceID: true}
	commander.Controller = game.Player2
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: commander.ObjectID, Target: game.AttackTarget{Player: game.Player3}},
		},
	}

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if got := g.Players[game.Player3].CommanderDamage[commander.CardInstanceID]; got != 7 {
		t.Fatalf("commander damage from stolen commander = %d, want 7", got)
	}
}

func TestTokenCopyOfCommanderDoesNotDealCommanderDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	commander := addCombatCreaturePermanentWithPower(g, game.Player1, 7)
	g.CommanderIDs = map[id.ID]bool{commander.CardInstanceID: true}
	card, ok := g.GetCardInstance(commander.CardInstanceID)
	if !ok {
		t.Fatal("commander card instance not found")
	}
	token, ok := createTokenPermanent(g, game.Player1, card.Def)
	if !ok {
		t.Fatal("token was not created")
	}
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: token.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if len(g.Players[game.Player2].CommanderDamage) != 0 {
		t.Fatalf("token copy commander damage = %+v, want none", g.Players[game.Player2].CommanderDamage)
	}
}

func TestCombatDamageUsesPowerCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pumped := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	pumped.Counters.Add(counter.PlusOnePlusOne, 1)
	shrunken := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	shrunken.Counters.Add(counter.MinusOneMinusOne, 1)
	zeroBase := addCombatCreaturePermanentWithPower(g, game.Player1, 0)
	zeroBase.Counters.Add(counter.PlusOnePlusOne, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: pumped.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: shrunken.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: zeroBase.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.resolveCombatDamage(g, &log)

	if g.Players[game.Player2].Life != 34 {
		t.Fatalf("defending player life = %d, want 34", g.Players[game.Player2].Life)
	}
	if len(log.CombatDamage) != 3 {
		t.Fatalf("combat damage logs = %d, want 3", len(log.CombatDamage))
	}
	if log.CombatDamage[0].Damage != 3 || log.CombatDamage[1].Damage != 1 || log.CombatDamage[2].Damage != 2 {
		t.Fatalf("combat damage = [%d %d %d], want [3 1 2]",
			log.CombatDamage[0].Damage, log.CombatDamage[1].Damage, log.CombatDamage[2].Damage)
	}
}

func TestBlockedCombatDamageMarksCreaturesAndPreventsPlayerDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
		},
	}
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.resolveCombatDamage(g, &log)

	if g.Players[game.Player2].Life != 40 {
		t.Fatalf("defending player life = %d, want 40", g.Players[game.Player2].Life)
	}
	if blocker.MarkedDamage != 3 {
		t.Fatalf("blocker marked damage = %d, want 3", blocker.MarkedDamage)
	}
	if attacker.MarkedDamage != 2 {
		t.Fatalf("attacker marked damage = %d, want 2", attacker.MarkedDamage)
	}
	if len(log.CombatDamage) != 0 {
		t.Fatalf("player combat damage logs = %+v, want none", log.CombatDamage)
	}
	if len(log.CreatureDamage) != 2 {
		t.Fatalf("creature damage logs = %d, want 2", len(log.CreatureDamage))
	}
}

func TestMultiBlockCombatDamageAssignsLethalDamageInOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: first.ObjectID, Blocking: attacker.ObjectID},
			{Blocker: second.ObjectID, Blocking: attacker.ObjectID},
		},
		BlockerOrder: map[id.ID][]id.ID{
			attacker.ObjectID: {first.ObjectID, second.ObjectID},
		},
	}
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.resolveCombatDamage(g, &log)

	if first.MarkedDamage != 2 {
		t.Fatalf("first blocker marked damage = %d, want 2", first.MarkedDamage)
	}
	if second.MarkedDamage != 3 {
		t.Fatalf("second blocker marked damage = %d, want 3", second.MarkedDamage)
	}
	if attacker.MarkedDamage != 4 {
		t.Fatalf("attacker marked damage = %d, want 4", attacker.MarkedDamage)
	}
	if g.Players[game.Player2].Life != 40 {
		t.Fatalf("defending player life = %d, want 40", g.Players[game.Player2].Life)
	}
	if len(log.CreatureDamage) != 4 {
		t.Fatalf("creature damage logs = %d, want 4", len(log.CreatureDamage))
	}
}

func TestMultiBlockCombatDamageStopsWhenInsufficientForFirstBlocker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 4)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: first.ObjectID, Blocking: attacker.ObjectID},
			{Blocker: second.ObjectID, Blocking: attacker.ObjectID},
		},
		BlockerOrder: map[id.ID][]id.ID{
			attacker.ObjectID: {first.ObjectID, second.ObjectID},
		},
	}
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.resolveCombatDamage(g, &log)

	if first.MarkedDamage != 3 {
		t.Fatalf("first blocker marked damage = %d, want 3", first.MarkedDamage)
	}
	if second.MarkedDamage != 0 {
		t.Fatalf("second blocker marked damage = %d, want 0", second.MarkedDamage)
	}
	if attacker.MarkedDamage != 6 {
		t.Fatalf("attacker marked damage = %d, want 6", attacker.MarkedDamage)
	}
}

func TestTrampleAssignsExcessDamageToDefendingPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 5, game.Trample)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
		},
	}
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.resolveCombatDamage(g, &log)

	if blocker.MarkedDamage != 2 {
		t.Fatalf("blocker marked damage = %d, want 2", blocker.MarkedDamage)
	}
	if g.Players[game.Player2].Life != 37 {
		t.Fatalf("defending player life = %d, want 37", g.Players[game.Player2].Life)
	}
	if len(log.CombatDamage) != 1 || log.CombatDamage[0].Damage != 3 {
		t.Fatalf("combat damage logs = %+v, want 3 trample damage", log.CombatDamage)
	}
}

func TestDeathtouchAssignsOneDamageAsLethal(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 5, game.Deathtouch)
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: first.ObjectID, Blocking: attacker.ObjectID},
			{Blocker: second.ObjectID, Blocking: attacker.ObjectID},
		},
		BlockerOrder: map[id.ID][]id.ID{
			attacker.ObjectID: {first.ObjectID, second.ObjectID},
		},
	}
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.resolveCombatDamage(g, &log)

	if first.MarkedDamage != 1 {
		t.Fatalf("first blocker marked damage = %d, want 1", first.MarkedDamage)
	}
	if !first.MarkedDeathtouchDamage {
		t.Fatal("first blocker did not record deathtouch damage")
	}
	if second.MarkedDamage != 4 {
		t.Fatalf("second blocker marked damage = %d, want 4", second.MarkedDamage)
	}
	if !second.MarkedDeathtouchDamage {
		t.Fatal("second blocker did not record deathtouch damage")
	}
	if g.Players[game.Player2].Life != 40 {
		t.Fatalf("defending player life = %d, want 40", g.Players[game.Player2].Life)
	}
}

func TestDeathtouchAssignsFreshDamageDespitePreexistingMarkedDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 5, game.Deathtouch)
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	first.MarkedDamage = 1
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: first.ObjectID, Blocking: attacker.ObjectID},
			{Blocker: second.ObjectID, Blocking: attacker.ObjectID},
		},
		BlockerOrder: map[id.ID][]id.ID{
			attacker.ObjectID: {first.ObjectID, second.ObjectID},
		},
	}
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.resolveCombatDamage(g, &log)

	if first.MarkedDamage != 2 {
		t.Fatalf("first blocker marked damage = %d, want 2", first.MarkedDamage)
	}
	if !first.MarkedDeathtouchDamage {
		t.Fatal("first blocker did not record deathtouch damage")
	}
	if second.MarkedDamage != 4 {
		t.Fatalf("second blocker marked damage = %d, want 4", second.MarkedDamage)
	}
}

func TestDeathtouchTrampleAssignsOneDamageBeforeTramplingOver(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 5, game.Deathtouch, game.Trample)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 10)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
		},
	}
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.resolveCombatDamage(g, &log)

	if blocker.MarkedDamage != 1 {
		t.Fatalf("blocker marked damage = %d, want 1", blocker.MarkedDamage)
	}
	if !blocker.MarkedDeathtouchDamage {
		t.Fatal("blocker did not record deathtouch damage")
	}
	if g.Players[game.Player2].Life != 36 {
		t.Fatalf("defending player life = %d, want 36", g.Players[game.Player2].Life)
	}
	if len(log.CombatDamage) != 1 || log.CombatDamage[0].Damage != 4 {
		t.Fatalf("combat damage logs = %+v, want 4 trample damage", log.CombatDamage)
	}
}

func TestDoubleStrikeTrampleDealsDamageWhenAllBlockersDieFirst(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.DoubleStrike, game.Trample)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.runCombatPhase(g, allFirstLegalAgents(), &log)

	if _, ok := permanentByObjectID(g, blocker.ObjectID); ok {
		t.Fatal("blocker survived first-strike trample damage")
	}
	if g.Players[game.Player2].Life != 36 {
		t.Fatalf("defending player life = %d, want 36", g.Players[game.Player2].Life)
	}
	if len(log.CombatDamage) != 2 {
		t.Fatalf("combat damage logs = %+v, want first-strike excess and normal trample damage", log.CombatDamage)
	}
	if log.CombatDamage[0].Damage != 1 || log.CombatDamage[1].Damage != 3 {
		t.Fatalf("combat damage amounts = [%d %d], want [1 3]", log.CombatDamage[0].Damage, log.CombatDamage[1].Damage)
	}
}

func TestFirstStrikeDeathtouchKillsBlockerBeforeNormalCombatDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 1, game.FirstStrike, game.Deathtouch)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.runCombatPhase(g, allFirstLegalAgents(), &log)

	if _, ok := permanentByObjectID(g, attacker.ObjectID); !ok {
		t.Fatal("first-strike deathtouch attacker died")
	}
	if attacker.MarkedDamage != 0 {
		t.Fatalf("attacker marked damage = %d, want 0", attacker.MarkedDamage)
	}
	if _, ok := permanentByObjectID(g, blocker.ObjectID); ok {
		t.Fatal("blocker survived first-strike deathtouch damage")
	}
	if len(log.Deaths) != 1 || log.Deaths[0].Permanent != blocker.ObjectID || log.Deaths[0].Reason != PermanentDeathReasonLethalDamage {
		t.Fatalf("death logs = %+v, want blocker lethal damage death", log.Deaths)
	}
}

func TestCombatWithFirstLegalBlockerKillsBlockedAttacker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.runCombatPhase(g, allFirstLegalAgents(), &log)

	if _, ok := permanentByObjectID(g, attacker.ObjectID); ok {
		t.Fatal("attacker survived lethal blocked combat damage")
	}
	if _, ok := permanentByObjectID(g, blocker.ObjectID); !ok {
		t.Fatal("blocker died despite nonlethal damage")
	}
	if !g.Players[game.Player1].Graveyard.Contains(attacker.CardInstanceID) {
		t.Fatal("dead attacker did not move to owner's graveyard")
	}
	if g.Players[game.Player2].Life != 40 {
		t.Fatalf("defending player life = %d, want 40", g.Players[game.Player2].Life)
	}
	if len(log.Deaths) != 1 || log.Deaths[0].Permanent != attacker.ObjectID || log.Deaths[0].Reason != PermanentDeathReasonLethalDamage {
		t.Fatalf("death logs = %+v, want attacker lethal damage death", log.Deaths)
	}
}

func TestFirstStrikeKillsBlockerBeforeNormalCombatDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2, game.FirstStrike)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.runCombatPhase(g, allFirstLegalAgents(), &log)

	if _, ok := permanentByObjectID(g, attacker.ObjectID); !ok {
		t.Fatal("first-strike attacker died")
	}
	if attacker.MarkedDamage != 0 {
		t.Fatalf("first-strike attacker marked damage = %d, want 0", attacker.MarkedDamage)
	}
	if _, ok := permanentByObjectID(g, blocker.ObjectID); ok {
		t.Fatal("blocker survived first-strike lethal damage")
	}
	if !g.Players[game.Player2].Graveyard.Contains(blocker.CardInstanceID) {
		t.Fatal("dead blocker did not move to graveyard")
	}
	if len(log.Deaths) != 1 || log.Deaths[0].Permanent != blocker.ObjectID {
		t.Fatalf("death logs = %+v, want blocker death", log.Deaths)
	}
}

func TestDoubleStrikeDealsDamageInBothCombatDamagePasses(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatCreaturePermanentWithPower(g, game.Player1, 2, game.DoubleStrike)
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.runCombatPhase(g, allFirstLegalAgents(), &log)

	if g.Players[game.Player2].Life != 36 {
		t.Fatalf("defending player life = %d, want 36", g.Players[game.Player2].Life)
	}
	if len(log.CombatDamage) != 2 {
		t.Fatalf("combat damage logs = %d, want 2", len(log.CombatDamage))
	}
	if log.CombatDamage[0].Damage != 2 || log.CombatDamage[1].Damage != 2 {
		t.Fatalf("combat damage amounts = [%d %d], want [2 2]", log.CombatDamage[0].Damage, log.CombatDamage[1].Damage)
	}
}

func TestCombatPhaseSkipsFirstStrikeStepWithoutFirstOrDoubleStrike(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	recorder := &combatStepRecorder{}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: recorder,
		game.Player2: recorder,
		game.Player3: recorder,
		game.Player4: recorder,
	}

	engine.runCombatPhase(g, agents, &TurnLog{})

	if slices.Contains(recorder.firstVisits, game.StepFirstStrikeDamage) {
		t.Fatalf("visited steps = %v, want no first-strike damage step", recorder.firstVisits)
	}
}

func TestCleanupStepClearsMarkedDamageOnSurvivors(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	survivor := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	survivor.MarkedDamage = 2
	survivor.MarkedDeathtouchDamage = true
	engine := NewEngine(nil)

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if survivor.MarkedDamage != 0 {
		t.Fatalf("marked damage after cleanup = %d, want 0", survivor.MarkedDamage)
	}
	if survivor.MarkedDeathtouchDamage {
		t.Fatal("marked deathtouch damage was not cleared during cleanup")
	}
}

func TestCleanupStepDiscardsActivePlayerToMaximumHandSize(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	var cards []id.ID
	for i := range 10 {
		cards = append(cards, addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: string(rune('A' + i))}}))
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if got := g.Players[game.Player1].Hand.Size(); got != maximumHandSize {
		t.Fatalf("hand size = %d, want %d", got, maximumHandSize)
	}
	for _, cardID := range cards[:3] {
		if !g.Players[game.Player1].Graveyard.Contains(cardID) {
			t.Fatalf("oldest overflow card %v was not discarded", cardID)
		}
	}
	for _, cardID := range cards[3:] {
		if !g.Players[game.Player1].Hand.Contains(cardID) {
			t.Fatalf("card %v should have remained in hand", cardID)
		}
	}
}

func TestCleanupStepDoesNotDiscardAtOrBelowMaximumHandSize(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for i := range maximumHandSize {
		addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: string(rune('A' + i))}})
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if got := g.Players[game.Player1].Hand.Size(); got != maximumHandSize {
		t.Fatalf("hand size = %d, want %d", got, maximumHandSize)
	}
	if got := g.Players[game.Player1].Graveyard.Size(); got != 0 {
		t.Fatalf("graveyard size = %d, want 0", got)
	}
}

func TestResolveCombatDamageNilAndStarPowerDealZero(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	nilPower := addCombatCreaturePermanent(g, game.Player1)
	starPower := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Star Creature",
		Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{IsStar: true})},
	})
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: nilPower.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: starPower.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.resolveCombatDamage(g, &log)

	if g.Players[game.Player2].Life != 40 {
		t.Fatalf("defending player life = %d, want 40", g.Players[game.Player2].Life)
	}
	if len(log.CombatDamage) != 0 {
		t.Fatalf("combat damage logs = %+v, want none", log.CombatDamage)
	}
}

func TestCombatDamageEliminatesPlayerBeforePostcombatMain(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Players[game.Player2].Life = 2
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.runCombatPhase(g, allFirstLegalAgents(), &log)

	if !g.Players[game.Player2].Eliminated {
		t.Fatal("defending player was not eliminated by combat damage")
	}
	if !g.TurnOrder.IsEliminated(game.Player2) {
		t.Fatal("defending player was not eliminated from turn order")
	}
	if g.Turn.Phase != game.PhaseCombat {
		t.Fatalf("phase = %v, want combat before postcombat main", g.Turn.Phase)
	}
	if len(log.Losses) != 1 || log.Losses[0].Player != game.Player2 || log.Losses[0].Reason != LossReasonZeroLife {
		t.Fatalf("loss logs = %+v, want Player2 0 life loss", log.Losses)
	}
}

type combatStepRecorder struct {
	firstVisits []game.Step
	seen        map[game.Step]bool
}

func (r *combatStepRecorder) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	if r.seen == nil {
		r.seen = make(map[game.Step]bool)
	}
	if obs.Turn.Phase == game.PhaseCombat && !r.seen[obs.Turn.Step] {
		r.seen[obs.Turn.Step] = true
		r.firstVisits = append(r.firstVisits, obs.Turn.Step)
	}
	return action.Pass()
}

type combatStateRecorder struct {
	game           *game.Game
	sawCombatState bool
}

func (r *combatStateRecorder) ChooseAction(obs PlayerObservation, legal []action.Action) action.Action {
	if obs.Turn.Phase == game.PhaseCombat && r.game.Combat != nil {
		r.sawCombatState = true
	}
	return action.Pass()
}

func addCombatCreaturePermanent(g *game.Game, controller game.PlayerID, keywords ...game.Keyword) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{Name: "Combat Creature",
		Types: []types.Card{
			types.Creature,
		},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: game.SimpleKeywords(keywords...),
		}}},
	})
}

func addCombatCreaturePermanentWithPower(g *game.Game, controller game.PlayerID, power int, keywords ...game.Keyword) *game.Permanent {
	pt := game.PT{Value: power}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{Name: "Powered Combat Creature",
		Types: []types.Card{
			types.Creature,
		},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: game.SimpleKeywords(keywords...),
		}}},
	})
}

func addCombatPermanent(g *game.Game, controller game.PlayerID, def *game.CardDef) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   def,
		Owner: controller,
	}
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          controller,
		Controller:     controller,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

func declareAttackersActionsContainTarget(actions []action.Action, attacker id.ID, target game.AttackTarget) bool {
	for _, act := range actions {
		payload, ok := act.DeclareAttackersPayload()
		if !ok {
			continue
		}
		for _, declaration := range payload.Attackers {
			if declaration.Attacker == attacker && declaration.Target == target {
				return true
			}
		}
	}
	return false
}

func mustDeclareAttackersPayload(t *testing.T, act action.Action) action.DeclareAttackersAction {
	t.Helper()
	payload, ok := act.DeclareAttackersPayload()
	if !ok {
		t.Fatalf("DeclareAttackersPayload() ok = false for %+v", act)
	}
	return payload
}

func mustDeclareBlockersPayload(t *testing.T, act action.Action) action.DeclareBlockersAction {
	t.Helper()
	payload, ok := act.DeclareBlockersPayload()
	if !ok {
		t.Fatalf("DeclareBlockersPayload() ok = false for %+v", act)
	}
	return payload
}

func intPtr(value int) *int {
	return new(value)
}

func permanentIDs(permanents []*game.Permanent) []id.ID {
	ids := make([]id.ID, 0, len(permanents))
	for _, permanent := range permanents {
		ids = append(ids, permanent.ObjectID)
	}
	return ids
}
