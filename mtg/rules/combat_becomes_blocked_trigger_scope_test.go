package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

// Source review: official CR effective 2026-09-25 (verified copy
// MagicCompRules 20260925.txt), read 509.1h-i, 509.3c-d, 509.3f and 702.25a-b.
// 509.3c: "Whenever [a creature] becomes blocked" generally triggers only once
// each combat for that creature, even if it's blocked by multiple creatures;
// 509.3d: "Whenever [a creature] becomes blocked by a creature" triggers once
// for each creature that blocks the specified creature; 702.25a: flanking means
// "Whenever this creature becomes blocked by a creature without flanking, the
// blocking creature gets -1/-1 until end of turn." Card faces (pinned corpus):
// Ashmouth Hound ({1}{R} 2/1 "Whenever this creature blocks or becomes blocked
// by a creature, this creature deals 1 damage to that creature."; Scryfall
// ruling 2011-09-22: it triggers once for each creature it blocks or becomes
// blocked by), Cave Tiger ({2}{G} 2/2 "Whenever this creature becomes blocked
// by a creature, this creature gets +1/+1 until end of turn."; no rulings),
// Zhalfirin Knight ({2}{W} 2/2 Flanking; "{W}{W}: This creature gains first
// strike until end of turn."; no rulings), Konda, Lord of Eiganjo ({5}{W}{W}
// 3/3 Vigilance, indestructible, Bushido 5 "Whenever this creature blocks or
// becomes blocked, it gets +5/+5 until end of turn."; no rulings). Retained
// responses: .cardwork/oracle-3494.json and .cardwork/rulings-3494.json in the
// unit worktree.

// doubleBlockedCombat attacks Player2 with attacker and has both blockers block
// it, resolving the whole combat.
func doubleBlockedCombat(t *testing.T, g *game.Game, attacker *game.Permanent, blockers ...*game.Permanent) *game.Game {
	t.Helper()
	attacker.SummoningSick = false
	declaration := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{
		attackAt(attacker, game.AttackTarget{Player: game.Player2}),
	}}
	defender := &bandBlockingAgent{}
	for _, blocker := range blockers {
		defender.blocks = append(defender.blocks, game.BlockDeclaration{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID})
	}
	policies := simPassPolicies()
	policies[game.Player2] = defender
	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, declaration, policies)
	if !ok {
		t.Fatal("combat not accepted")
	}
	if got := len(blockEvents(after.Events, game.EventAttackerBecameBlocked)); got != len(blockers) {
		t.Fatalf("became-blocked pairs = %d, want %d (one per blocking creature)", got, len(blockers))
	}
	return after
}

// TestBecomesBlockedByACreatureTriggerFiresPerBlocker runs registered Ashmouth
// Hound blocked by two 4/4s: its "becomes blocked by a creature" trigger fires
// for each blocker (CR 509.3d), so each takes 1 damage before combat damage
// adds the Hound's 2, 4 marked in all.
func TestBecomesBlockedByACreatureTriggerFiresPerBlocker(t *testing.T) {
	g := bandCombatGame()
	hound := registryCombatPermanent(t, g, game.Player1, "Ashmouth Hound")
	first := addCombatPermanent(g, game.Player2, vanillaCreature("First Wall", 4, 4))
	second := addCombatPermanent(g, game.Player2, vanillaCreature("Second Wall", 4, 4))

	after := doubleBlockedCombat(t, g, hound, first, second)
	total := 0
	for _, blocker := range []*game.Permanent{first, second} {
		afterBlocker, ok := permanentByObjectID(after, blocker.ObjectID)
		if !ok {
			t.Fatalf("blocker %v left the battlefield", blocker.ObjectID)
		}
		if afterBlocker.MarkedDamage < 1 {
			t.Fatalf("blocker %v marked damage = %d, want at least 1 from Ashmouth Hound's trigger", blocker.ObjectID, afterBlocker.MarkedDamage)
		}
		total += afterBlocker.MarkedDamage
	}
	if total != 4 {
		t.Fatalf("total marked damage on blockers = %d, want 4 (two triggers plus 2 combat damage)", total)
	}
}

// TestBecomesBlockedByACreatureSelfPumpFiresPerBlocker runs registered Cave
// Tiger blocked by two 0/5s: it gets +1/+1 for each blocker (CR 509.3d), 4/4.
func TestBecomesBlockedByACreatureSelfPumpFiresPerBlocker(t *testing.T) {
	g := bandCombatGame()
	tiger := registryCombatPermanent(t, g, game.Player1, "Cave Tiger")
	first := addCombatPermanent(g, game.Player2, vanillaCreature("First Wall", 0, 5))
	second := addCombatPermanent(g, game.Player2, vanillaCreature("Second Wall", 0, 5))

	after := doubleBlockedCombat(t, g, tiger, first, second)
	afterTiger, ok := permanentByObjectID(after, tiger.ObjectID)
	if !ok {
		t.Fatal("Cave Tiger left the battlefield")
	}
	if got := effectivePower(after, afterTiger); got != 4 {
		t.Fatalf("Cave Tiger power = %d, want 4: +1/+1 per blocking creature (CR 509.3d)", got)
	}
}

// TestFlankingGivesEachNonFlankingBlockerMinusOne runs registered Zhalfirin
// Knight blocked by two 1/5s: flanking triggers for each blocker without
// flanking (CR 702.25a), so both are 0/4.
func TestFlankingGivesEachNonFlankingBlockerMinusOne(t *testing.T) {
	g := bandCombatGame()
	knight := registryCombatPermanent(t, g, game.Player1, "Zhalfirin Knight")
	first := addCombatPermanent(g, game.Player2, vanillaCreature("First Wall", 1, 5))
	second := addCombatPermanent(g, game.Player2, vanillaCreature("Second Wall", 1, 5))

	after := doubleBlockedCombat(t, g, knight, first, second)
	for _, blocker := range []*game.Permanent{first, second} {
		afterBlocker, ok := permanentByObjectID(after, blocker.ObjectID)
		if !ok {
			t.Fatalf("blocker %v left the battlefield", blocker.ObjectID)
		}
		toughness, _ := effectiveToughness(after, afterBlocker)
		if power := effectivePower(after, afterBlocker); power != 0 || toughness != 4 {
			t.Fatalf("blocker %v is %d/%d, want 0/4: flanking gives each non-flanking blocker -1/-1", blocker.ObjectID, power, toughness)
		}
	}
	if _, ok := permanentByObjectID(after, knight.ObjectID); !ok {
		t.Fatal("Zhalfirin Knight died to 0-power blockers")
	}
}

// TestBecomesBlockedTriggerFiresOnceForTwoBlockers runs registered Konda, Lord
// of Eiganjo blocked by two creatures: the "becomes blocked" half of bushido 5
// fires once (CR 509.3c), so Konda is 8/8, not 13/13.
func TestBecomesBlockedTriggerFiresOnceForTwoBlockers(t *testing.T) {
	g := bandCombatGame()
	konda := registryCombatPermanent(t, g, game.Player1, "Konda, Lord of Eiganjo")
	first := addCombatPermanent(g, game.Player2, vanillaCreature("First Wall", 1, 9))
	second := addCombatPermanent(g, game.Player2, vanillaCreature("Second Wall", 1, 9))

	after := doubleBlockedCombat(t, g, konda, first, second)
	afterKonda, ok := permanentByObjectID(after, konda.ObjectID)
	if !ok {
		t.Fatal("Konda left the battlefield")
	}
	if got := effectivePower(after, afterKonda); got != 8 {
		t.Fatalf("Konda power = %d, want 8: bushido's becomes-blocked trigger fires once (CR 509.3c)", got)
	}
}
