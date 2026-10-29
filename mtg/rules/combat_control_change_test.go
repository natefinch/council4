package rules

import (
	"slices"
	"testing"

	cardc "github.com/natefinch/council4/mtg/cards/c"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3496. Official CR effective 2026-09-25 (verified local
// copy), read 506.4, 509.1g-h, 510.1c-d and 802.4a. 506.4: a permanent is
// removed from combat if its controller or protector changes; a creature
// removed from combat stops being an attacking, blocking, blocked and/or
// unblocked creature, and a planeswalker removed from combat stops being
// attacked. 506.4c: a creature attacking a planeswalker removed from combat
// continues attacking, attacking nothing; unblocked, it deals no combat damage.
// 509.1g: a blocking creature remains blocking until it's removed from combat.
// 509.1h: a creature remains blocked even if all its blockers are removed from
// combat. 510.1c-d: a blocked creature with no blockers assigns no combat
// damage (absent trample). Cards (pinned corpus): Word of Seizing (oracle
// b80d2841-03b0-42d3-ac9b-51e793b1dc24, {3}{R}{R} instant) "Split second ...
// Untap target permanent and gain control of it until end of turn. It gains
// haste until end of turn." (rulings 2021-03-19 cover split second only).
// Control Magic (oracle cd0d7141-46d2-4aa3-bc77-6b3b4513803e, {2}{U}{U} Aura)
// "Enchant creature / You control enchanted creature." (ruling 2016-06-08:
// gaining control of a creature doesn't gain control of Auras or Equipment
// attached to it). Disenchant (oracle a7e97fa9-4b72-4548-b854-5be5f18a6f1a,
// {1}{W} instant) "Destroy target artifact or enchantment." Grizzly Bears
// (oracle 14c8f55d-d177-4c25-a931-ebeb9e6062a0, {1}{G} 2/2).

// castAndResolveWordOfSeizing has caster cast the registered Word of Seizing on
// target with priority and resolves it.
func castAndResolveWordOfSeizing(t *testing.T, g *game.Game, engine *Engine, caster game.PlayerID, target *game.Permanent) {
	t.Helper()
	castAndResolveRegistered(t, g, engine, caster, "Word of Seizing", mana.R, 5, target)
	if effectiveController(g, target) != caster {
		t.Fatalf("Word of Seizing left %s under player %d", permanentName(g, target), effectiveController(g, target))
	}
}

func castAndResolveRegistered(t *testing.T, g *game.Game, engine *Engine, caster game.PlayerID, name string, color mana.Color, amount int, target *game.Permanent) {
	t.Helper()
	spell := addCardToHand(g, caster, registryCardDef(t, name))
	g.Players[caster].ManaPool.Add(color, amount)
	g.Turn.PriorityPlayer = caster
	cast := action.CastSpell(spell, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)
	if !containsAction(engine.legalActions(g, caster), cast) {
		t.Fatalf("%s could not target %s", name, permanentName(g, target))
	}
	if !engine.applyAction(g, caster, cast) {
		t.Fatalf("casting %s failed", name)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
}

// assertNoCombatDamage resolves combat damage and checks that no player lost
// life and that none of the creatures was dealt damage.
func assertNoCombatDamage(t *testing.T, g *game.Game, engine *Engine, creatures ...*game.Permanent) {
	t.Helper()
	var life []int
	for _, player := range g.Players {
		life = append(life, player.Life)
	}
	engine.resolveCombatDamage(g, &TurnLog{})
	for i, player := range g.Players {
		if player.Life != life[i] {
			t.Fatalf("player %d life = %d, want %d", i, player.Life, life[i])
		}
	}
	for _, creature := range creatures {
		if damage := markedDamage(t, g, creature); damage != 0 {
			t.Fatalf("%s was dealt %d combat damage, want none", permanentName(g, creature), damage)
		}
	}
}

// TestWordOfSeizingStealsBlockedAttackerOutOfCombat is the #3496 scenario. The
// defending player casts the registered Word of Seizing on a blocked attacker.
// The attacker's controller changes, so it is removed from combat: it is no
// longer attacking or blocked, deals and is dealt no combat damage, and its
// blocker remains a blocking creature (CR 506.4, 509.1g).
func TestWordOfSeizingStealsBlockedAttackerOutOfCombat(t *testing.T) {
	g := bandCombatGame()
	engine := choicesProgramEngine(t)
	attacker := registryCombatPermanent(t, g, game.Player1, "Grizzly Bears")
	blocker := registryCombatPermanent(t, g, game.Player2, "Grizzly Bears")
	blockedAttacker(t, g, engine, attacker, blocker)

	castAndResolveWordOfSeizing(t, g, engine, game.Player2, attacker)
	if attacker.Tapped {
		t.Fatal("Word of Seizing did not untap the stolen attacker")
	}
	if len(g.Combat.Attackers) != 0 {
		t.Fatalf("attackers = %+v, want the stolen creature removed from combat", g.Combat.Attackers)
	}
	assertRemovedAttackerLeftBlockerBlocking(t, g, attacker, blocker)
	if _, ok := g.Combat.CombatantControllers[attacker.ObjectID]; ok {
		t.Fatal("the removed attacker kept its combatant control record")
	}
	assertNoCombatDamage(t, g, engine, attacker, blocker)
}

// TestWordOfSeizingStealsBlockerOutOfCombat has the attacking player steal the
// creature blocking its attacker. The blocker is removed from combat and stops
// blocking, but the attacker remains blocked (CR 509.1h), so it deals no
// combat damage to the defending player and neither creature is dealt any.
func TestWordOfSeizingStealsBlockerOutOfCombat(t *testing.T) {
	g := bandCombatGame()
	engine := choicesProgramEngine(t)
	attacker := registryCombatPermanent(t, g, game.Player1, "Grizzly Bears")
	blocker := registryCombatPermanent(t, g, game.Player2, "Grizzly Bears")
	blockedAttacker(t, g, engine, attacker, blocker)

	castAndResolveWordOfSeizing(t, g, engine, game.Player1, blocker)
	if combatStateMatches(g, blocker, game.CombatStateAttackingOrBlocking) || len(g.Combat.Blockers) != 0 {
		t.Fatalf("the stolen blocker is still in combat (blocks %+v)", g.Combat.Blockers)
	}
	if !combatStateMatches(g, attacker, game.CombatStateAttacking) || !attackerWasBlocked(g, attacker.ObjectID) {
		t.Fatal("the attacker stopped attacking or became unblocked when its blocker was stolen")
	}
	assertNoCombatDamage(t, g, engine, attacker, blocker)
}

// TestControlMagicEndingRemovesBlockerFromCombat ends a static control effect
// mid-combat. A creature Player2 controls through Control Magic blocks; Player1
// casts the registered Disenchant on the Aura, so the creature returns to its
// owner's control and is removed from combat. The attacker remains blocked and
// deals no combat damage.
func TestControlMagicEndingRemovesBlockerFromCombat(t *testing.T) {
	g := bandCombatGame()
	engine := choicesProgramEngine(t)
	attacker := registryCombatPermanent(t, g, game.Player1, "Grizzly Bears")
	blocker := registryCombatPermanent(t, g, game.Player1, "Grizzly Bears")
	aura := addCombatPermanent(g, game.Player2, cardc.ControlMagic())
	if !attachPermanent(g, aura, blocker) || effectiveController(g, blocker) != game.Player2 {
		t.Fatal("fixture did not give Player2 control of the enchanted creature")
	}
	blockedAttacker(t, g, engine, attacker, blocker)

	castAndResolveRegistered(t, g, engine, game.Player1, "Disenchant", mana.W, 2, aura)
	if _, present := permanentByObjectID(g, aura.ObjectID); present {
		t.Fatal("Disenchant did not destroy Control Magic")
	}
	if effectiveController(g, blocker) != game.Player1 {
		t.Fatal("the creature did not return to its owner's control")
	}
	if combatStateMatches(g, blocker, game.CombatStateAttackingOrBlocking) || len(g.Combat.Blockers) != 0 {
		t.Fatalf("the creature whose control effect ended is still blocking (blocks %+v)", g.Combat.Blockers)
	}
	if !attackerWasBlocked(g, attacker.ObjectID) {
		t.Fatal("the attacker became unblocked when its blocker was removed from combat")
	}
	assertNoCombatDamage(t, g, engine, attacker, blocker)
}

// TestControlEffectStartingIsSettledBeforePriority starts a static control
// effect without a resolving spell or zone change: Control Magic becomes
// attached to an attacking creature. The state-based-action check that
// precedes priority settles the control change and removes the creature from
// combat, while its blocker keeps blocking.
func TestControlEffectStartingIsSettledBeforePriority(t *testing.T) {
	g := bandCombatGame()
	engine := choicesProgramEngine(t)
	attacker := registryCombatPermanent(t, g, game.Player1, "Grizzly Bears")
	blocker := registryCombatPermanent(t, g, game.Player2, "Grizzly Bears")
	blockedAttacker(t, g, engine, attacker, blocker)
	aura := addCombatPermanent(g, game.Player2, cardc.ControlMagic())
	if !attachPermanent(g, aura, attacker) || effectiveController(g, attacker) != game.Player2 {
		t.Fatal("fixture did not give Player2 control of the attacker")
	}
	engine.applyStateBasedActions(g)
	if len(g.Combat.Attackers) != 0 {
		t.Fatalf("attackers = %+v, want the stolen creature removed from combat", g.Combat.Attackers)
	}
	assertRemovedAttackerLeftBlockerBlocking(t, g, attacker, blocker)
	assertNoCombatDamage(t, g, engine, attacker, blocker)
}

// TestStolenPlaneswalkerStopsBeingAttacked has the attacking player gain
// control of the planeswalker its creature attacks. The planeswalker is removed
// from combat and stops being attacked; the creature keeps attacking, attacking
// nothing, so unblocked it deals no combat damage (CR 506.4, 506.4c).
func TestStolenPlaneswalkerStopsBeingAttacked(t *testing.T) {
	g := bandCombatGame()
	engine := choicesProgramEngine(t)
	attacker := registryCombatPermanent(t, g, game.Player1, "Grizzly Bears")
	planeswalker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:    "Test Planeswalker",
		Types:   []types.Card{types.Planeswalker},
		Loyalty: opt.Val(3),
	}})
	planeswalker.Counters.Add(counter.Loyalty, 3)
	target := game.AttackTarget{Player: game.Player2, PlaneswalkerID: planeswalker.ObjectID}
	if !engine.applyDeclareAttackers(g, game.Player1, action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{attackAt(attacker, target)}}) {
		t.Fatal("attack on the planeswalker was rejected")
	}
	g.Turn.Step = game.StepDeclareBlockers
	if !engine.applyDeclareBlockers(g, game.Player2, action.DeclareBlockersAction{}) {
		t.Fatal("empty block declaration was rejected")
	}

	castAndResolveWordOfSeizing(t, g, engine, game.Player1, planeswalker)
	want := []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{NoTarget: true, Player: game.Player2}}}
	if !slices.Equal(g.Combat.Attackers, want) {
		t.Fatalf("attackers = %+v, want %+v", g.Combat.Attackers, want)
	}
	assertNoCombatDamage(t, g, engine, attacker)
	if loyalty := planeswalker.Counters.Get(counter.Loyalty); loyalty != 3 {
		t.Fatalf("planeswalker loyalty = %d, want 3", loyalty)
	}
}
