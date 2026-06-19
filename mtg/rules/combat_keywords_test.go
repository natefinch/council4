package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

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

func TestWitherDamageAddsMinusOneMinusOneCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.Wither)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	g.Combat = blockedCombat(attacker, blocker)

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if got := blocker.Counters.Get(counter.MinusOneMinusOne); got != 3 {
		t.Fatalf("blocker -1/-1 counters = %d, want 3", got)
	}
	if blocker.MarkedDamage != 0 {
		t.Fatalf("blocker marked damage = %d, want 0", blocker.MarkedDamage)
	}
}

func TestWitherDamageKillsWhenToughnessReachesZero(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2, game.Wither)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Combat = blockedCombat(attacker, blocker)
	engine := NewEngine(nil)

	engine.resolveCombatDamage(g, &TurnLog{})
	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if _, ok := permanentByObjectID(g, blocker.ObjectID); ok {
		t.Fatal("creature with toughness reduced to 0 remained on battlefield")
	}
	if _, ok := permanentByObjectID(g, attacker.ObjectID); ok {
		t.Fatal("Wither attacker survived simultaneous lethal blocker damage")
	}
	if !slices.ContainsFunc(deaths, func(death PermanentDeathLog) bool {
		return death.Permanent == blocker.ObjectID && death.Reason == PermanentDeathReasonZeroToughness
	}) {
		t.Fatalf("deaths = %+v, want blocker death from 0 toughness", deaths)
	}
	if !slices.ContainsFunc(deaths, func(death PermanentDeathLog) bool {
		return death.Permanent == attacker.ObjectID && death.Reason == PermanentDeathReasonLethalDamage
	}) {
		t.Fatalf("deaths = %+v, want attacker death from lethal damage", deaths)
	}
}

func TestWitherDamageCountersCancelWithPlusOnePlusOne(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2, game.Wither)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	blocker.Counters.Add(counter.PlusOnePlusOne, 2)
	g.Combat = blockedCombat(attacker, blocker)
	engine := NewEngine(nil)

	engine.resolveCombatDamage(g, &TurnLog{})
	engine.applyStateBasedActions(g)

	if got := blocker.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("blocker +1/+1 counters = %d, want 0", got)
	}
	if got := blocker.Counters.Get(counter.MinusOneMinusOne); got != 0 {
		t.Fatalf("blocker -1/-1 counters = %d, want 0", got)
	}
	if blocker.MarkedDamage != 0 {
		t.Fatalf("blocker marked damage = %d, want 0", blocker.MarkedDamage)
	}
}

func TestNonWitherDamageStillUsesMarkedDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	g.Combat = blockedCombat(attacker, blocker)

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if blocker.MarkedDamage != 3 {
		t.Fatalf("blocker marked damage = %d, want 3", blocker.MarkedDamage)
	}
	if got := blocker.Counters.Get(counter.MinusOneMinusOne); got != 0 {
		t.Fatalf("blocker -1/-1 counters = %d, want 0", got)
	}
}

func TestWitherGrantedByContinuousEffectApplies(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Wither Granter",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility,
				Group: game.ObjectControlledGroup(
					game.SourcePermanentReference(),
					game.Selection{RequiredTypes: []types.Card{types.Creature}},
				),
				AddKeywords: []game.Keyword{game.Wither},
			}},
		}},
	}})
	g.Combat = blockedCombat(attacker, blocker)

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if got := blocker.Counters.Get(counter.MinusOneMinusOne); got != 3 {
		t.Fatalf("blocker -1/-1 counters = %d, want 3", got)
	}
	if blocker.MarkedDamage != 0 {
		t.Fatalf("blocker marked damage = %d, want 0", blocker.MarkedDamage)
	}
}

// TestInfectGrantedByContinuousEffectApplies proves that infect granted to an
// attacker by a static keyword-grant continuous effect (the "<subject> has
// infect" declaration) routes combat damage to creatures as -1/-1 counters,
// exactly like printed infect.
func TestInfectGrantedByContinuousEffectApplies(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Infect Granter",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility,
				Group: game.ObjectControlledGroup(
					game.SourcePermanentReference(),
					game.Selection{RequiredTypes: []types.Card{types.Creature}},
				),
				AddKeywords: []game.Keyword{game.Infect},
			}},
		}},
	}})
	g.Combat = blockedCombat(attacker, blocker)

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if got := blocker.Counters.Get(counter.MinusOneMinusOne); got != 3 {
		t.Fatalf("blocker -1/-1 counters = %d, want 3", got)
	}
	if blocker.MarkedDamage != 0 {
		t.Fatalf("blocker marked damage = %d, want 0", blocker.MarkedDamage)
	}
}

func TestInfectDamageToCreatureAddsMinusOneMinusOneCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3, game.Infect)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	g.Combat = blockedCombat(attacker, blocker)

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if got := blocker.Counters.Get(counter.MinusOneMinusOne); got != 3 {
		t.Fatalf("blocker -1/-1 counters = %d, want 3", got)
	}
	if blocker.MarkedDamage != 0 {
		t.Fatalf("blocker marked damage = %d, want 0", blocker.MarkedDamage)
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

func TestToxicAddsPoisonAfterCombatDamageToPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Toxic Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			{KeywordAbilities: []game.KeywordAbility{game.ToxicKeyword{Amount: 1}}},
			{KeywordAbilities: []game.KeywordAbility{game.ToxicKeyword{Amount: 2}}},
		},
	}})

	markPlayerCombatDamage(g, source, game.Player2, 2, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != 38 {
		t.Fatalf("life = %d, want 38", got)
	}
	if got := g.Players[game.Player2].PoisonCounters; got != 3 {
		t.Fatalf("poison counters = %d, want 3", got)
	}
}
