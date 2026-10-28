package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/opt"
)

// Source review: official CR effective 2026-09-25 (verified copy
// MagicCompRules 20260925.txt, SHA256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca), read
// 510.1-510.2, 702.19b, 702.2c and 702.22j. 702.19b: "When checking for
// assigned lethal damage, take into account damage already marked on the
// creature and damage from other creatures that's being assigned during the
// same combat damage step, but not any abilities or effects that might change
// the amount of damage that's actually dealt", with the example of a 2/2 that
// blocks a 1/1 and a 3/3 trampler being assigned 1 by each and 2 trampling
// over. 702.2c: any nonzero combat damage assigned by a deathtouch source is
// lethal for determining excess damage. 510.1e: the total assignment of a
// player's creatures is checked, so it can't depend on the order attackers
// were declared. 702.22j: with a banding blocker the defending player chooses
// the trampler's assignment under the same rules. Scryfall Oracle retrieved
// 2026-10-28: War Elephant 2b7c9fe0-5a23-4172-b31f-b0d85cd465f6 "Trample;
// banding", Kjeldoran Escort 1c6c1604-0724-40b1-9df6-71e57e5063c9 2/3
// "Banding", Colossal Dreadmaw 08c7db90-c0cf-4482-b7ee-bb033e5996d2 6/6
// "Trample", Typhoid Rats d6ee6cc1-902d-4f56-afa5-6fa4813bfbbc 1/1
// "Deathtouch", Grizzly Bears 14c8f55d-d177-4c25-a931-ebeb9e6062a0 2/2
// vanilla. Rulings for War Elephant and Kjeldoran Escort (2009-10-01 "If a
// creature in combat has banding, its controller assigns damage for creatures
// blocking or blocked by it.") agree; the others have none.

// sharedBlockerCombat has blocker block every attacker, each attacking
// Player2, with the attackers declared in the given order.
func sharedBlockerCombat(g *game.Game, blocker *game.Permanent, attackers ...*game.Permanent) *game.CombatState {
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind: game.RuleEffectCanBlockAdditional, SourceObjectID: blocker.ObjectID, AffectedSource: true, AdditionalBlockCount: len(attackers) - 1,
	})
	combat := &game.CombatState{}
	for _, attacker := range attackers {
		combat.Attackers = append(combat.Attackers, game.AttackDeclaration{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}})
		combat.Blockers = append(combat.Blockers, game.BlockDeclaration{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID})
	}
	return combat
}

// TestTrampleCountsSimultaneousAssignmentCRExample plays the CR 702.19b
// example: a 2/2 that can block an additional creature blocks a 1/1 and a 3/3
// trampler. The 1/1's 1 damage counts toward lethal, so the trampler assigns
// 1 to the blocker and 2 to the defending player, in either declaration order.
func TestTrampleCountsSimultaneousAssignmentCRExample(t *testing.T) {
	for _, tramplerFirst := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		small := bandingTestCreature(g, game.Player1, 1, 1)
		trampler := bandingTestCreature(g, game.Player1, 3, 3, game.Trample)
		blocker := bandingTestCreature(g, game.Player2, 2, 2)
		attackers := []*game.Permanent{small, trampler}
		if tramplerFirst {
			attackers = []*game.Permanent{trampler, small}
		}
		g.Combat = sharedBlockerCombat(g, blocker, attackers...)

		NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

		if blocker.MarkedDamage != 2 || g.Players[game.Player2].Life != 38 {
			t.Fatalf("trampler first %v: blocker damage/defender life = %d/%d, want 2/38", tramplerFirst, blocker.MarkedDamage, g.Players[game.Player2].Life)
		}
	}
}

// TestTrampleCountsSimultaneousAssignmentRegistryCards is the non-banding
// case with registered cards: Grizzly Bears and Colossal Dreadmaw attack and a
// 4/4 that can block an additional creature blocks both. The Bears' 2 counts
// toward lethal, so the Dreadmaw assigns 2 to the blocker and tramples 4 over,
// whichever attacker was declared first.
func TestTrampleCountsSimultaneousAssignmentRegistryCards(t *testing.T) {
	for _, dreadmawFirst := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		bears := registryCombatPermanent(t, g, game.Player1, "Grizzly Bears")
		dreadmaw := registryCombatPermanent(t, g, game.Player1, "Colossal Dreadmaw")
		blocker := bandingTestCreature(g, game.Player2, 4, 4)
		attackers := []*game.Permanent{bears, dreadmaw}
		if dreadmawFirst {
			attackers = []*game.Permanent{dreadmaw, bears}
		}
		g.Combat = sharedBlockerCombat(g, blocker, attackers...)
		engine := NewEngine(nil)

		engine.resolveCombatDamage(g, &TurnLog{})
		engine.applyStateBasedActions(g)

		if blocker.MarkedDamage != 4 || g.Players[game.Player2].Life != 36 {
			t.Fatalf("Dreadmaw first %v: blocker damage/defender life = %d/%d, want 4/36", dreadmawFirst, blocker.MarkedDamage, g.Players[game.Player2].Life)
		}
		if permanentWithCardOnBattlefield(g, blocker.CardInstanceID) {
			t.Fatalf("Dreadmaw first %v: the 4/4 was dealt lethal damage and should die", dreadmawFirst)
		}
	}
}

// TestTrampleChosenAssignmentCountsSimultaneousAssignment: a recorded
// assignment of only 2 of Colossal Dreadmaw's 6 to a 4/4 with 4 trampling
// over is legal because Grizzly Bears assigns the other 2 in the same step, so
// it is honored rather than replaced by the default.
func TestTrampleChosenAssignmentCountsSimultaneousAssignment(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	dreadmaw := registryCombatPermanent(t, g, game.Player1, "Colossal Dreadmaw")
	bears := registryCombatPermanent(t, g, game.Player1, "Grizzly Bears")
	blocker := bandingTestCreature(g, game.Player2, 4, 4)
	g.Combat = sharedBlockerCombat(g, blocker, dreadmaw, bears)
	g.Combat.DamageAssignment = map[id.ID]int{blocker.ObjectID: 2}

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if blocker.MarkedDamage != 4 || g.Players[game.Player2].Life != 36 {
		t.Fatalf("blocker damage/defender life = %d/%d, want 4/36", blocker.MarkedDamage, g.Players[game.Player2].Life)
	}
}

// TestTrampleCountsSimultaneousDeathtouchAssignment: Typhoid Rats' 1 deathtouch
// damage assigned to a 7/7 in the same step is lethal (CR 702.2c), so Colossal
// Dreadmaw needs to assign nothing to the 7/7 and tramples all 6 over.
func TestTrampleCountsSimultaneousDeathtouchAssignment(t *testing.T) {
	for _, dreadmawFirst := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		rats := registryCombatPermanent(t, g, game.Player1, "Typhoid Rats")
		dreadmaw := registryCombatPermanent(t, g, game.Player1, "Colossal Dreadmaw")
		blocker := bandingTestCreature(g, game.Player2, 7, 7)
		attackers := []*game.Permanent{rats, dreadmaw}
		if dreadmawFirst {
			attackers = []*game.Permanent{dreadmaw, rats}
		}
		g.Combat = sharedBlockerCombat(g, blocker, attackers...)
		engine := NewEngine(nil)

		engine.resolveCombatDamage(g, &TurnLog{})
		engine.applyStateBasedActions(g)

		if blocker.MarkedDamage != 1 || g.Players[game.Player2].Life != 34 {
			t.Fatalf("Dreadmaw first %v: blocker damage/defender life = %d/%d, want 1/34", dreadmawFirst, blocker.MarkedDamage, g.Players[game.Player2].Life)
		}
		if permanentWithCardOnBattlefield(g, blocker.CardInstanceID) {
			t.Fatalf("Dreadmaw first %v: the 7/7 was dealt deathtouch damage and should die", dreadmawFirst)
		}
	}
}

// TestTwoTramplersShareLethalAssignment: Colossal Dreadmaw and a 3/3 trampler
// both blocked by a 5/5 together assign exactly lethal damage to it, each
// counting the other's assignment, so 4 tramples over in either order.
func TestTwoTramplersShareLethalAssignment(t *testing.T) {
	for _, dreadmawFirst := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		dreadmaw := registryCombatPermanent(t, g, game.Player1, "Colossal Dreadmaw")
		small := bandingTestCreature(g, game.Player1, 3, 3, game.Trample)
		blocker := bandingTestCreature(g, game.Player2, 5, 5)
		attackers := []*game.Permanent{small, dreadmaw}
		if dreadmawFirst {
			attackers = []*game.Permanent{dreadmaw, small}
		}
		g.Combat = sharedBlockerCombat(g, blocker, attackers...)

		NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

		if blocker.MarkedDamage != 5 || g.Players[game.Player2].Life != 36 {
			t.Fatalf("Dreadmaw first %v: blocker damage/defender life = %d/%d, want 5/36", dreadmawFirst, blocker.MarkedDamage, g.Players[game.Player2].Life)
		}
	}
}

// TestTrampleCountsAssignedDamageThatWillBePrevented: an effect that changes
// the damage actually dealt is not counted (CR 702.19b). The 4/4 blocker's
// shield counter will prevent the damage, but Grizzly Bears' assigned 2 still
// counts, so Colossal Dreadmaw assigns only 2 to it and tramples 4 over.
func TestTrampleCountsAssignedDamageThatWillBePrevented(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	bears := registryCombatPermanent(t, g, game.Player1, "Grizzly Bears")
	dreadmaw := registryCombatPermanent(t, g, game.Player1, "Colossal Dreadmaw")
	blocker := bandingTestCreature(g, game.Player2, 4, 4)
	blocker.Counters.Add(counter.Shield, 1)
	g.Combat = sharedBlockerCombat(g, blocker, dreadmaw, bears)

	NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

	if g.Players[game.Player2].Life != 36 {
		t.Fatalf("defender life = %d, want 36 (4 trampled over)", g.Players[game.Player2].Life)
	}
	if blocker.MarkedDamage != 0 || blocker.Counters.Get(counter.Shield) != 0 {
		t.Fatalf("blocker damage/shield = %d/%d, want the shield to prevent the 4 assigned", blocker.MarkedDamage, blocker.Counters.Get(counter.Shield))
	}
}

// TestBandingTrampleExcessCountsSimultaneousAssignment plays the issue example
// through the Simulator: registered War Elephant (2/2 trample, banding) and
// Grizzly Bears attack, and registered Kjeldoran Escort (2/3 banding), able to
// block an additional creature, blocks both. The defending player divides both
// attackers' damage (CR 702.22j); the Bears' 2 counts toward the Escort's
// lethal damage, so they may assign 1 of War Elephant's damage to the Escort
// and 1 to themself, in either declaration order. Their default keeps it all
// on the Escort.
func TestBandingTrampleExcessCountsSimultaneousAssignment(t *testing.T) {
	for _, tc := range []struct {
		name          string
		elephantFirst bool
		excess        int
		life          int
	}{
		{name: "excess to the defender, Bears declared first", excess: 1, life: 39},
		{name: "excess to the defender, War Elephant declared first", elephantFirst: true, excess: 1, life: 39},
		{name: "default keeps the damage on the Escort", excess: 0, life: 40},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			elephant := registryCombatPermanent(t, g, game.Player1, "War Elephant")
			bears := registryCombatPermanent(t, g, game.Player1, "Grizzly Bears")
			escort := registryCombatPermanent(t, g, game.Player2, "Kjeldoran Escort")
			g.RuleEffects = append(g.RuleEffects, game.RuleEffect{Kind: game.RuleEffectCanBlockAdditional, SourceObjectID: escort.ObjectID, AffectedSource: true, AdditionalBlockCount: 1})
			g.Turn.ActivePlayer = game.Player1
			g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
			g.Combat = &game.CombatState{}
			attackerAgent := &bandingDivisionAgent{}
			defender := &bandingDivisionAgent{}
			if tc.excess > 0 {
				defender.answer = trampleExcessAnswer(tc.excess, nil)
			}
			policies := simPassPolicies()
			policies[game.Player1], policies[game.Player2] = attackerAgent, defender
			declarations := []game.AttackDeclaration{
				{Attacker: bears.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
				{Attacker: elephant.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			}
			if tc.elephantFirst {
				declarations[0], declarations[1] = declarations[1], declarations[0]
			}

			after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, action.DeclareAttackersAction{Attackers: declarations}, policies)
			if !ok {
				t.Fatal("combat not accepted")
			}

			if len(defender.requests) != 1 {
				t.Fatalf("defender requests = %+v, want only War Elephant's trample excess", defender.requests)
			}
			number := defender.requests[0]
			if number.Kind != game.ChoiceNumber || number.CombatAttacker != elephant.ObjectID ||
				!number.NumberRange.Exists || number.NumberRange.Val.Maximum != opt.Val(1) {
				t.Fatalf("trample excess request = %+v, want War Elephant's 0..1", number)
			}
			if got := after.Players[game.Player2].Life; got != tc.life {
				t.Fatalf("defending player life = %d, want %d", got, tc.life)
			}
			if permanentWithCardOnBattlefield(after, escort.CardInstanceID) {
				t.Fatal("Kjeldoran Escort was assigned lethal damage and should die")
			}
			if after.Execution.InvalidChoices != 0 {
				t.Fatalf("invalid choices = %d, want 0", after.Execution.InvalidChoices)
			}
		})
	}
}

// sharedBlockersCombat has every blocker block every attacker, each attacking
// Player2, with the attackers declared in the given order.
func sharedBlockersCombat(g *game.Game, blockers []*game.Permanent, attackers ...*game.Permanent) *game.CombatState {
	combat := &game.CombatState{}
	for _, attacker := range attackers {
		combat.Attackers = append(combat.Attackers, game.AttackDeclaration{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}})
	}
	for _, blocker := range blockers {
		g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
			Kind: game.RuleEffectCanBlockAdditional, SourceObjectID: blocker.ObjectID, AffectedSource: true, AdditionalBlockCount: len(attackers) - 1,
		})
		for _, attacker := range attackers {
			combat.Blockers = append(combat.Blockers, game.BlockDeclaration{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID})
		}
	}
	return combat
}

// TestTrampleJointAssignmentAcrossSharedBlockers: Colossal Dreadmaw and a 4/4
// trampler are each blocked by two 3/3s. Their 10 damage must jointly assign
// lethal damage, 6, to the blockers, so 4 tramples over in either declaration
// order: whichever assigns first relies on the other's damage, and the other
// must then assign what it was relied on for.
func TestTrampleJointAssignmentAcrossSharedBlockers(t *testing.T) {
	for _, dreadmawFirst := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		dreadmaw := registryCombatPermanent(t, g, game.Player1, "Colossal Dreadmaw")
		other := bandingTestCreature(g, game.Player1, 4, 4, game.Trample)
		first := bandingTestCreature(g, game.Player2, 3, 3)
		second := bandingTestCreature(g, game.Player2, 3, 3)
		attackers := []*game.Permanent{other, dreadmaw}
		if dreadmawFirst {
			attackers = []*game.Permanent{dreadmaw, other}
		}
		g.Combat = sharedBlockersCombat(g, []*game.Permanent{first, second}, attackers...)
		engine := NewEngine(nil)

		engine.resolveCombatDamage(g, &TurnLog{})
		engine.applyStateBasedActions(g)

		if first.MarkedDamage != 3 || second.MarkedDamage != 3 || g.Players[game.Player2].Life != 36 {
			t.Fatalf("Dreadmaw first %v: blocker damage %d/%d, defender life %d, want 3/3 and 36", dreadmawFirst, first.MarkedDamage, second.MarkedDamage, g.Players[game.Player2].Life)
		}
		if permanentWithCardOnBattlefield(g, first.CardInstanceID) || permanentWithCardOnBattlefield(g, second.CardInstanceID) {
			t.Fatalf("Dreadmaw first %v: both blockers were assigned lethal damage and should die", dreadmawFirst)
		}
	}
}

// TestTrampleDeathtouchTramplerJointAssignment: Typhoid Rats granted trample
// and Colossal Dreadmaw are both blocked by a 7/7. The Rats' 1 deathtouch
// damage is lethal (CR 702.2c), so Dreadmaw tramples all 6 over, in either
// declaration order, though the Rats assign after it when declared later.
func TestTrampleDeathtouchTramplerJointAssignment(t *testing.T) {
	for _, dreadmawFirst := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		dreadmaw := registryCombatPermanent(t, g, game.Player1, "Colossal Dreadmaw")
		rats := registryCombatPermanent(t, g, game.Player1, "Typhoid Rats")
		g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
			Layer: game.LayerAbility, AffectedObjectID: rats.ObjectID, AddKeywords: []game.Keyword{game.Trample},
		})
		blocker := bandingTestCreature(g, game.Player2, 7, 7)
		attackers := []*game.Permanent{rats, dreadmaw}
		if dreadmawFirst {
			attackers = []*game.Permanent{dreadmaw, rats}
		}
		g.Combat = sharedBlockerCombat(g, blocker, attackers...)

		NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

		if blocker.MarkedDamage != 1 || g.Players[game.Player2].Life != 34 {
			t.Fatalf("Dreadmaw first %v: blocker damage/defender life = %d/%d, want 1/34", dreadmawFirst, blocker.MarkedDamage, g.Players[game.Player2].Life)
		}
	}
}

// TestTrampleRecordedAssignmentRelyingOnLaterTrampler: Colossal Dreadmaw and a
// 3/3 trampler share a 5/5 blocker, with a recorded assignment of 2 to it. The
// first to assign may assign only 2 because the other's damage completes
// lethal damage; that commits the other to assign the 3 still needed, so its
// recorded 2 isn't legal and it assigns all 3. 4 tramples over either way.
func TestTrampleRecordedAssignmentRelyingOnLaterTrampler(t *testing.T) {
	for _, dreadmawFirst := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		dreadmaw := registryCombatPermanent(t, g, game.Player1, "Colossal Dreadmaw")
		small := bandingTestCreature(g, game.Player1, 3, 3, game.Trample)
		blocker := bandingTestCreature(g, game.Player2, 5, 5)
		attackers := []*game.Permanent{small, dreadmaw}
		if dreadmawFirst {
			attackers = []*game.Permanent{dreadmaw, small}
		}
		g.Combat = sharedBlockerCombat(g, blocker, attackers...)
		g.Combat.DamageAssignment = map[id.ID]int{blocker.ObjectID: 2}

		NewEngine(nil).resolveCombatDamage(g, &TurnLog{})

		if blocker.MarkedDamage != 5 || g.Players[game.Player2].Life != 36 {
			t.Fatalf("Dreadmaw first %v: blocker damage/defender life = %d/%d, want 5/36", dreadmawFirst, blocker.MarkedDamage, g.Players[game.Player2].Life)
		}
	}
}

// TestBandingTrampleJointExcessIndependentOfOrder: registered Colossal
// Dreadmaw and War Elephant are both blocked by registered Kjeldoran Escort,
// so the defending player assigns both (CR 702.22j). The Elephant's 2 can
// count toward the Escort's lethal damage, so the defending player may
// trample up to 5 of Dreadmaw's damage over in either declaration order. When
// Dreadmaw assigns first and relies on it, the Elephant must assign its 2 to
// the Escort and nothing is asked about it.
func TestBandingTrampleJointExcessIndependentOfOrder(t *testing.T) {
	for _, dreadmawFirst := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		dreadmaw := registryCombatPermanent(t, g, game.Player1, "Colossal Dreadmaw")
		elephant := registryCombatPermanent(t, g, game.Player1, "War Elephant")
		escort := registryCombatPermanent(t, g, game.Player2, "Kjeldoran Escort")
		attackers := []*game.Permanent{elephant, dreadmaw}
		if dreadmawFirst {
			attackers = []*game.Permanent{dreadmaw, elephant}
		}
		g.Combat = sharedBlockerCombat(g, escort, attackers...)
		defender := &bandingDivisionAgent{answer: func(request game.ChoiceRequest) []int {
			if request.Kind == game.ChoiceNumber && request.CombatAttacker == dreadmaw.ObjectID {
				return []int{5}
			}
			return request.DefaultSelection
		}}
		var agents [game.NumPlayers]PlayerAgent
		agents[game.Player2] = defender

		combatEngine{NewEngine(nil)}.resolveDamagePassWithAgents(g, normalCombatDamage, agents, &TurnLog{})

		dreadmawMax, elephantAsked := -1, false
		for _, request := range defender.requests {
			switch {
			case request.Kind == game.ChoiceNumber && request.CombatAttacker == dreadmaw.ObjectID:
				dreadmawMax = request.NumberRange.Val.Maximum.Val
			case request.CombatAttacker == elephant.ObjectID:
				elephantAsked = true
			}
		}
		if dreadmawMax != 5 {
			t.Fatalf("Dreadmaw first %v: Dreadmaw's trample excess maximum = %d, want 5", dreadmawFirst, dreadmawMax)
		}
		if dreadmawFirst && elephantAsked {
			t.Fatalf("Dreadmaw first: the Elephant's assignment was asked, but it must assign its 2 to the Escort")
		}
		if g.Players[game.Player2].Life != 35 || escort.MarkedDamage != 3 {
			t.Fatalf("Dreadmaw first %v: defender life/Escort damage = %d/%d, want 35/3", dreadmawFirst, g.Players[game.Player2].Life, escort.MarkedDamage)
		}
	}
}

// TestBandingTrampleChoosesLeastJointAssignment: registered Colossal Dreadmaw
// and War Elephant attack and two registered Kjeldoran Escorts each block
// both. Dreadmaw assigns first; with the Elephant's 2 still to come it must
// assign at least 4 to the Escorts to trample 2 over, which it may divide 3/1,
// 2/2 or 1/3. The defending player chooses 1/3, so the Elephant must assign
// its 2 to the first Escort.
func TestBandingTrampleChoosesLeastJointAssignment(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	dreadmaw := registryCombatPermanent(t, g, game.Player1, "Colossal Dreadmaw")
	elephant := registryCombatPermanent(t, g, game.Player1, "War Elephant")
	first := registryCombatPermanent(t, g, game.Player2, "Kjeldoran Escort")
	second := registryCombatPermanent(t, g, game.Player2, "Kjeldoran Escort")
	g.Combat = sharedBlockersCombat(g, []*game.Permanent{first, second}, dreadmaw, elephant)
	defender := &bandingDivisionAgent{answer: func(request game.ChoiceRequest) []int {
		switch {
		case request.Kind == game.ChoiceNumber && request.CombatAttacker == dreadmaw.ObjectID:
			return []int{2}
		case request.Kind == game.ChoiceModal && request.CombatAttacker == dreadmaw.ObjectID:
			return []int{2}
		}
		return request.DefaultSelection
	}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player2] = defender

	combatEngine{NewEngine(nil)}.resolveDamagePassWithAgents(g, normalCombatDamage, agents, &TurnLog{})

	var labels []string
	for _, request := range defender.requests {
		if request.CombatAttacker == elephant.ObjectID {
			t.Fatalf("the Elephant's assignment was asked (%+v), but it must assign its 2 to the first Escort", request)
		}
		if request.Kind == game.ChoiceNumber && request.NumberRange.Val.Maximum != opt.Val(2) {
			t.Fatalf("Dreadmaw's trample excess maximum = %v, want 2", request.NumberRange.Val.Maximum)
		}
		if request.Kind == game.ChoiceModal {
			for _, option := range request.Options {
				labels = append(labels, option.Label)
			}
		}
	}
	if len(labels) != 3 {
		t.Fatalf("least assignment options = %q, want 3/1, 2/2 and 1/3", labels)
	}
	if first.MarkedDamage != 3 || second.MarkedDamage != 3 || g.Players[game.Player2].Life != 38 {
		t.Fatalf("Escort damage %d/%d, defender life %d, want 3/3 and 38", first.MarkedDamage, second.MarkedDamage, g.Players[game.Player2].Life)
	}
}
