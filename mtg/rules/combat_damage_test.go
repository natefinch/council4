package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

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
