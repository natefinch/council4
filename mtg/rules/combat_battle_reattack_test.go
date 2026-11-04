package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: issue #3597. Official CR effective 2026-09-25 (verified local
// copy MagicCompRules20260925.txt), read 506.4, 506.4c, 508.4, 310.9d,
// 702.49a and 702.49c. 506.4: a permanent is removed from combat if its
// controller or protector changes; a battle removed from combat stops being
// attacked. 506.4c: a creature attacking a removed battle keeps attacking
// nothing and, unblocked, deals no combat damage. 508.4: a creature put onto
// the battlefield attacking attacks the chosen or specified battle a defending
// player protects. 702.49a/c: ninjutsu returns an unblocked attacker as a cost
// and the Ninja enters attacking the same player, planeswalker or battle. So a
// control change while a battle isn't attacked doesn't affect a later attack on
// it; only a change while it is attacked removes it. Cards (pinned corpus):
// Invasion of Segovia (oracle 7b3c0d89-55ed-43d7-a00a-7b060857ec4e, Battle —
// Siege, defense 4; rulings 2023-04-14: if its controller changes during
// combat it is removed from combat). Okiba-Gang Shinobi (oracle
// efe3b58f-6f83-45d5-beb1-7dde1cae9dd7, {3}{B}{B} 3/2) "Ninjutsu {3}{B} ({3}{B},
// Return an unblocked attacker you control to hand: Put this card onto the
// battlefield from your hand tapped and attacking.) Whenever this creature
// deals combat damage to a player, that player discards two cards." Word of
// Seizing (oracle b80d2841-03b0-42d3-ac9b-51e793b1dc24, {3}{R}{R} instant)
// "Split second ... Untap target permanent and gain control of it until end of
// turn. It gains haste until end of turn." Grizzly Bears (oracle
// 14c8f55d-d177-4c25-a931-ebeb9e6062a0, {1}{G} 2/2).

// ninjaAttacksSiegeStolenWhileUnattacked is the #3597 scenario. Player1's
// Grizzly Bears attack Player2's Invasion of Segovia protected by Player3, who
// declares no blocks. Player1 activates Okiba-Gang Shinobi's ninjutsu,
// returning the Bears, so the Siege is no longer attacked; in response Player4
// resolves Word of Seizing on it, leaving its protector unchanged. The ninjutsu
// then resolves and control is settled.
func ninjaAttacksSiegeStolenWhileUnattacked(t *testing.T) (*game.Game, *Engine, *game.Permanent, *game.Permanent) {
	t.Helper()
	g, engine, bears, siege := siegeUnderAttack(t, game.Player2, game.Player3)
	g.Turn.Step = game.StepDeclareBlockers
	if !engine.applyDeclareBlockers(g, game.Player3, action.DeclareBlockersAction{}) {
		t.Fatal("the protector's empty block declaration was rejected")
	}
	ninjaID := addCardToHand(g, game.Player1, registryCardDef(t, "Okiba-Gang Shinobi"))
	g.Players[game.Player1].ManaPool.Add(mana.B, 4)
	g.Turn.PriorityPlayer = game.Player1
	agents := simPassPolicies()
	log := &TurnLog{}
	activate := action.ActivateAbility(ninjaID, 0, nil, 0)
	if !actionsContain(engine.legalActions(g, game.Player1), activate) {
		t.Fatal("Okiba-Gang Shinobi ninjutsu is not a legal action")
	}
	if !engine.applyActionWithChoices(g, game.Player1, activate, agents, log) {
		t.Fatal("Okiba-Gang Shinobi ninjutsu activation failed")
	}
	if !g.Players[game.Player1].Hand.Contains(bears.CardInstanceID) || len(g.Combat.Attackers) != 0 {
		t.Fatalf("attackers after ninjutsu cost = %+v, want the Bears returned and no attackers", g.Combat.Attackers)
	}
	if _, recorded := g.Combat.CombatantControllers[siege.ObjectID]; recorded {
		t.Fatal("the unattacked Siege kept its combat controller record")
	}

	castAndResolveWordOfSeizing(t, g, engine, game.Player4, siege)
	if protector, ok := battleProtector(siege); !ok || protector != game.Player3 {
		t.Fatalf("protector = %v/%v, want Player3", protector, ok)
	}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	settleControlChanges(g)

	ninja := permanentWithCardID(g, ninjaID)
	if ninja == nil || !ninja.Tapped {
		t.Fatalf("Okiba-Gang Shinobi = %+v, want it on the battlefield tapped", ninja)
	}
	return g, engine, ninja, siege
}

// TestNinjaAttacksSiegeStolenWhileUnattacked checks that the Ninja attacks the
// Siege stolen while no creature attacked it: the Siege's earlier controller
// record was discarded, its current controller is recorded, and the unblocked
// Ninja deals combat damage to it.
func TestNinjaAttacksSiegeStolenWhileUnattacked(t *testing.T) {
	g, engine, ninja, siege := ninjaAttacksSiegeStolenWhileUnattacked(t)
	want := []game.AttackDeclaration{{Attacker: ninja.ObjectID, Target: game.AttackTarget{Player: game.Player3, BattleID: siege.ObjectID}}}
	if !slices.Equal(g.Combat.Attackers, want) {
		t.Fatalf("attackers = %+v, want %+v", g.Combat.Attackers, want)
	}
	if recorded, ok := g.Combat.CombatantControllers[siege.ObjectID]; !ok || recorded != game.Player4 {
		t.Fatalf("Siege combat controller = %v/%v, want Player4", recorded, ok)
	}
	engine.resolveCombatDamage(g, &TurnLog{})
	if defense := siege.Counters.Get(counter.Defense); defense != 1 {
		t.Fatalf("Siege defense = %d, want 1 after the Ninja's 3 combat damage", defense)
	}
}

// TestSiegeStolenWhileNinjaAttacksStopsBeingAttacked continues the #3597
// scenario: Player2 regains control of the Siege while the Ninja attacks it, so
// the Siege is removed from combat and the Ninja attacks nothing (CR 506.4,
// 506.4c).
func TestSiegeStolenWhileNinjaAttacksStopsBeingAttacked(t *testing.T) {
	g, engine, ninja, siege := ninjaAttacksSiegeStolenWhileUnattacked(t)
	castAndResolveWordOfSeizing(t, g, engine, game.Player2, siege)
	want := []game.AttackDeclaration{{Attacker: ninja.ObjectID, Target: game.AttackTarget{NoTarget: true, Player: game.Player3}}}
	if !slices.Equal(g.Combat.Attackers, want) {
		t.Fatalf("attackers = %+v, want %+v", g.Combat.Attackers, want)
	}
	if _, recorded := g.Combat.CombatantControllers[siege.ObjectID]; recorded {
		t.Fatal("the Siege removed from combat kept its combat controller record")
	}
	assertNoCombatDamage(t, g, engine, ninja)
	if defense := siege.Counters.Get(counter.Defense); defense != 4 {
		t.Fatalf("Siege defense = %d, want 4", defense)
	}
}

// TestTokenJoiningAttackOnRestolenSiegeRecordsCurrentController uses the
// put-onto-the-battlefield-attacking-a-battle operation (mobilize's route). The
// Bears attacking the Siege are stolen out of combat, leaving it unattacked;
// Player4 then steals the Siege. A creature put onto the battlefield attacking
// the Siege records Player4, keeps attacking it through settlement, and the
// Siege is removed from combat only when its controller changes again.
func TestTokenJoiningAttackOnRestolenSiegeRecordsCurrentController(t *testing.T) {
	g, engine, bears, siege := siegeUnderAttack(t, game.Player2, game.Player3)
	target := game.AttackTarget{Player: game.Player3, BattleID: siege.ObjectID}
	castAndResolveWordOfSeizing(t, g, engine, game.Player2, bears)
	if len(g.Combat.Attackers) != 0 {
		t.Fatalf("attackers = %+v, want the stolen Bears removed from combat", g.Combat.Attackers)
	}
	if _, recorded := g.Combat.CombatantControllers[siege.ObjectID]; recorded {
		t.Fatal("the unattacked Siege kept its combat controller record")
	}
	castAndResolveWordOfSeizing(t, g, engine, game.Player4, siege)

	entering := registryCombatPermanent(t, g, game.Player1, "Grizzly Bears")
	declareTokenAttackingTarget(g, game.Player1, entering, target)
	if recorded, ok := g.Combat.CombatantControllers[siege.ObjectID]; !ok || recorded != game.Player4 {
		t.Fatalf("Siege combat controller = %v/%v, want Player4 recorded as it became attacked", recorded, ok)
	}
	settleControlChanges(g)
	want := []game.AttackDeclaration{{Attacker: entering.ObjectID, Target: target}}
	if !slices.Equal(g.Combat.Attackers, want) {
		t.Fatalf("attackers = %+v, want %+v", g.Combat.Attackers, want)
	}

	castAndResolveWordOfSeizing(t, g, engine, game.Player1, siege)
	want = []game.AttackDeclaration{{Attacker: entering.ObjectID, Target: game.AttackTarget{NoTarget: true, Player: game.Player3}}}
	if !slices.Equal(g.Combat.Attackers, want) {
		t.Fatalf("attackers = %+v, want %+v", g.Combat.Attackers, want)
	}
}
