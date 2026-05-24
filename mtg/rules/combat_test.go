package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
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
	nonCreature := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:  "Relic",
		Types: []game.CardType{game.TypeArtifact},
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

	if len(legal) != 3 {
		t.Fatalf("legal declare attackers actions = %d, want 3", len(legal))
	}
	wantTargets := []game.PlayerID{game.Player2, game.Player4}
	for i, target := range wantTargets {
		if legal[i].Kind != action.ActionDeclareAttackers {
			t.Fatalf("action %d kind = %v, want declare attackers", i, legal[i].Kind)
		}
		want := []game.AttackDeclaration{
			{Attacker: attacker1.ObjectID, Target: game.AttackTarget{Player: target}},
			{Attacker: attacker2.ObjectID, Target: game.AttackTarget{Player: target}},
		}
		if !slices.Equal(legal[i].DeclareAttackers.Attackers, want) {
			t.Fatalf("action %d attackers = %+v, want %+v", i, legal[i].DeclareAttackers.Attackers, want)
		}
	}
	if len(legal[2].DeclareAttackers.Attackers) != 0 {
		t.Fatalf("last declare attackers action = %+v, want no attacks", legal[2].DeclareAttackers.Attackers)
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
	declare := action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: normal.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: vigilance.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}).DeclareAttackers

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
				return action.DeclareAttackers([]game.AttackDeclaration{
					{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
					{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player3}},
				}).DeclareAttackers
			},
		},
		{
			name: "dead defending player",
			declare: func(g *game.Game, attacker *game.Permanent) action.DeclareAttackersAction {
				g.Players[game.Player2].Eliminated = true
				g.TurnOrder.Eliminate(game.Player2)
				return action.DeclareAttackers([]game.AttackDeclaration{
					{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
				}).DeclareAttackers
			},
		},
		{
			name: "planeswalker target",
			declare: func(g *game.Game, attacker *game.Permanent) action.DeclareAttackersAction {
				return action.DeclareAttackers([]game.AttackDeclaration{
					{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2, PlaneswalkerID: 99}},
				}).DeclareAttackers
			},
		},
		{
			name: "summoning sick attacker",
			declare: func(g *game.Game, attacker *game.Permanent) action.DeclareAttackersAction {
				attacker.SummoningSick = true
				return action.DeclareAttackers([]game.AttackDeclaration{
					{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
				}).DeclareAttackers
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
	if len(legal[1].DeclareBlockers.Blockers) != 0 {
		t.Fatalf("last block action = %+v, want no blocks", legal[1].DeclareBlockers.Blockers)
	}
	for _, act := range legal {
		for _, block := range act.DeclareBlockers.Blockers {
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
	declare := action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
	}).DeclareBlockers

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
				return action.DeclareBlockers([]game.BlockDeclaration{
					{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
					{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID + 100},
				}).DeclareBlockers
			},
		},
		{
			name: "duplicate attacker",
			declare: func(g *game.Game, attacker *game.Permanent, blocker *game.Permanent) action.DeclareBlockersAction {
				other := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
				return action.DeclareBlockers([]game.BlockDeclaration{
					{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
					{Blocker: other.ObjectID, Blocking: attacker.ObjectID},
				}).DeclareBlockers
			},
		},
		{
			name: "tapped blocker",
			declare: func(g *game.Game, attacker *game.Permanent, blocker *game.Permanent) action.DeclareBlockersAction {
				blocker.Tapped = true
				return action.DeclareBlockers([]game.BlockDeclaration{
					{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
				}).DeclareBlockers
			},
		},
		{
			name: "attacker not attacking controller",
			declare: func(g *game.Game, attacker *game.Permanent, blocker *game.Permanent) action.DeclareBlockersAction {
				g.Combat.Attackers[0].Target.Player = game.Player3
				return action.DeclareBlockers([]game.BlockDeclaration{
					{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID},
				}).DeclareBlockers
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

func TestCombatWithFirstLegalBlockerKillsBlockedAttacker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.runCombatPhase(g, allFirstLegalAgents(), &log)

	if permanentByObjectID(g, attacker.ObjectID) != nil {
		t.Fatal("attacker survived lethal blocked combat damage")
	}
	if permanentByObjectID(g, blocker.ObjectID) == nil {
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

func TestCleanupStepClearsMarkedDamageOnSurvivors(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	survivor := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	survivor.MarkedDamage = 2
	engine := NewEngine(nil)

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if survivor.MarkedDamage != 0 {
		t.Fatalf("marked damage after cleanup = %d, want 0", survivor.MarkedDamage)
	}
}

func TestResolveCombatDamageNilAndStarPowerDealZero(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	nilPower := addCombatCreaturePermanent(g, game.Player1)
	starPower := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:  "Star Creature",
		Types: []game.CardType{game.TypeCreature},
		Power: &game.PT{IsStar: true},
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
	return addCombatPermanent(g, controller, &game.CardDef{
		Name: "Combat Creature",
		Types: []game.CardType{
			game.TypeCreature,
		},
		Abilities: []game.AbilityDef{
			{
				Kind:     game.StaticAbility,
				Keywords: keywords,
			},
		},
	})
}

func addCombatCreaturePermanentWithPower(g *game.Game, controller game.PlayerID, power int, keywords ...game.Keyword) *game.Permanent {
	pt := game.PT{Value: power}
	return addCombatPermanent(g, controller, &game.CardDef{
		Name: "Powered Combat Creature",
		Types: []game.CardType{
			game.TypeCreature,
		},
		Power:     &pt,
		Toughness: &pt,
		Abilities: []game.AbilityDef{
			{
				Kind:     game.StaticAbility,
				Keywords: keywords,
			},
		},
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

func permanentIDs(permanents []*game.Permanent) []id.ID {
	ids := make([]id.ID, 0, len(permanents))
	for _, permanent := range permanents {
		ids = append(ids, permanent.ObjectID)
	}
	return ids
}
