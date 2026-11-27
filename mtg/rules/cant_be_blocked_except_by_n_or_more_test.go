package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

// Source review: Comprehensive Rules effective 2026-11-06 (retained copy
// .cardwork/cr-20261106.txt in the unit worktree). 509.1b: a block
// declaration that disobeys any restriction is illegal. 509.1c: requirements
// are maximized without disobeying restrictions. 702.111b: "A creature with
// menace can't be blocked except by two or more creatures." 702.23a-b: Rampage
// N gives +N/+N for each creature blocking it beyond the first, calculated once
// when the trigger resolves. Pinned-corpus and Scryfall Oracle (retained
// .cardwork/card-<oracleId>.json): Rampaging Ceratops ({4}{R} 5/4), Pathrazer
// of Ulamog ({11} 9/9, Annihilator 3), Gorilla Berserkers ({3}{G}{G} 2/3,
// "Trample; rampage 2"), Phyrexian Colossus ({7} 8/8), Guile ({3}{U}{U}{U}
// 6/6) and Underworld Cerberus ({3}{B}{R} 6/6) each read "This creature can't
// be blocked except by three or more creatures." No ruling addresses the
// minimum itself; Pathrazer's 2010-06-15 rulings only place annihilator's
// sacrifices before blocks. These scenarios use rules fixtures with each
// card's restriction; the cards' other abilities are card-authoring work.

// exceptByThreeOrMore builds a creature fixture carrying "This creature can't
// be blocked except by three or more creatures."
func exceptByThreeOrMore(name string, power, toughness int, keywords ...game.Keyword) *game.CardDef {
	def := vanillaCreature(name, power, toughness, keywords...)
	def.StaticAbilities = append(def.StaticAbilities, game.StaticAbility{
		Text: "This creature can't be blocked except by three or more creatures.",
		RuleEffects: []game.RuleEffect{{
			Kind:            game.RuleEffectCantBeBlockedExceptByNOrMore,
			AffectedSource:  true,
			MinimumBlockers: 3,
		}},
	})
	return def
}

// threeOrMoreAttack returns a declare-blockers game in which Player1's fixture
// attacks Player2, who controls blockerCount untapped 1/1s.
func threeOrMoreAttack(def *game.CardDef, blockerCount int) (*game.Game, *game.Permanent, []*game.Permanent) {
	g := mustBlockGame()
	attacker := addCombatPermanent(g, game.Player1, def)
	attackingDefender(g, game.Player2, attacker)
	var blockers []*game.Permanent
	for range blockerCount {
		blockers = append(blockers, addCombatPermanent(g, game.Player2, vanillaCreature("Footsoldier", 1, 1)))
	}
	return g, attacker, blockers
}

func blocksBy(attacker *game.Permanent, blockers ...*game.Permanent) []game.BlockDeclaration {
	declarations := make([]game.BlockDeclaration, 0, len(blockers))
	for _, blocker := range blockers {
		declarations = append(declarations, game.BlockDeclaration{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID})
	}
	return declarations
}

// offeredBlockCounts returns the number of creatures blocking attacker in each
// offered declare-blockers sample.
func offeredBlockCounts(t *testing.T, g *game.Game, attacker *game.Permanent) []int {
	t.Helper()
	var counts []int
	for _, act := range (combatEngine{}).legalBlockers(g, game.Player2) {
		payload := mustDeclareBlockersPayload(t, act)
		count := 0
		for _, block := range payload.Blockers {
			if block.Blocking == attacker.ObjectID {
				count++
			}
		}
		counts = append(counts, count)
	}
	return counts
}

// TestRampagingCeratopsTwoBlockersCannotBlock: with only two potential
// blockers Rampaging Ceratops can't be blocked. No block is offered, and a
// two-creature declaration is rejected.
func TestRampagingCeratopsTwoBlockersCannotBlock(t *testing.T) {
	g, ceratops, blockers := threeOrMoreAttack(exceptByThreeOrMore("Rampaging Ceratops", 5, 4), 2)
	for _, count := range offeredBlockCounts(t, g, ceratops) {
		if count != 0 {
			t.Fatalf("offered a %d-creature block of Rampaging Ceratops, want none", count)
		}
	}
	checkBlockLegality(t, g, map[string]struct {
		blocks []game.BlockDeclaration
		legal  bool
	}{
		"no block": {nil, true},
		"one":      {blocksBy(ceratops, blockers[0]), false},
		"both":     {blocksBy(ceratops, blockers...), false},
	})
	if NewEngine(nil).applyDeclareBlockers(g, game.Player2, action.DeclareBlockersAction{Blockers: blocksBy(ceratops, blockers...)}) {
		t.Fatal("two blockers were applied to Rampaging Ceratops")
	}
	if len(g.Combat.Blockers) != 0 {
		t.Fatalf("blockers = %+v, want none after the rejected declaration", g.Combat.Blockers)
	}
}

// TestRampagingCeratopsThreeOrMoreBlockers: three blockers may block Rampaging
// Ceratops, and with four potential blockers any three or all four are legal
// while one or two are not.
func TestRampagingCeratopsThreeOrMoreBlockers(t *testing.T) {
	g, ceratops, blockers := threeOrMoreAttack(exceptByThreeOrMore("Rampaging Ceratops", 5, 4), 3)
	offered := offeredBlockCounts(t, g, ceratops)
	for _, count := range offered {
		if count != 0 && count < 3 {
			t.Fatalf("offered a %d-creature block of Rampaging Ceratops", count)
		}
	}
	if !slices.Contains(offered, 3) {
		t.Fatalf("offered block counts %v, want the three-creature block", offered)
	}
	if !NewEngine(nil).applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, blocksOf(ceratops, blockers...))) {
		t.Fatal("three blockers were rejected for Rampaging Ceratops")
	}
	if len(g.Combat.Blockers) != 3 {
		t.Fatalf("blockers = %+v, want all three", g.Combat.Blockers)
	}

	g, ceratops, blockers = threeOrMoreAttack(exceptByThreeOrMore("Rampaging Ceratops", 5, 4), 4)
	checkBlockLegality(t, g, map[string]struct {
		blocks []game.BlockDeclaration
		legal  bool
	}{
		"one":      {blocksBy(ceratops, blockers[0]), false},
		"two":      {blocksBy(ceratops, blockers[1], blockers[3]), false},
		"three":    {blocksBy(ceratops, blockers[0], blockers[2], blockers[3]), true},
		"all four": {blocksBy(ceratops, blockers...), true},
		"no block": {nil, true},
	})
}

// TestGorillaBerserkersBlockedByThreeRampages plays Gorilla Berserkers' text
// through a full combat: three creatures block it (the fewest it allows),
// rampage 2 gives +4/+4 for the two blockers beyond the first, and the 6/7
// tramples 3 over three 1/1s.
func TestGorillaBerserkersBlockedByThreeRampages(t *testing.T) {
	def := exceptByThreeOrMore("Gorilla Berserkers", 2, 3, game.Trample)
	def.TriggeredAbilities = append(def.TriggeredAbilities, game.RampageTriggeredAbility(2))
	g := bandCombatGame()
	berserkers := addCombatPermanent(g, game.Player1, def)
	berserkers.SummoningSick = false
	var blockers []*game.Permanent
	for range 3 {
		blockers = append(blockers, addCombatPermanent(g, game.Player2, vanillaCreature("Footsoldier", 1, 1)))
	}
	life := g.Players[game.Player2].Life

	defender := &bandBlockingAgent{blocks: blocksBy(berserkers, blockers...)}
	policies := simPassPolicies()
	policies[game.Player2] = defender
	declaration := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{
		attackAt(berserkers, game.AttackTarget{Player: game.Player2}),
	}}
	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, declaration, policies)
	if !ok {
		t.Fatal("combat not accepted")
	}
	if got := len(blockEvents(after.Events, game.EventAttackerBecameBlocked)); got != 3 {
		t.Fatalf("became-blocked pairs = %d, want 3", got)
	}
	afterBerserkers, ok := permanentByObjectID(after, berserkers.ObjectID)
	if !ok {
		t.Fatal("Gorilla Berserkers left the battlefield")
	}
	power := effectivePower(after, afterBerserkers)
	if toughness, _ := effectiveToughness(after, afterBerserkers); power != 6 || toughness != 7 {
		t.Fatalf("Gorilla Berserkers = %d/%d, want 6/7 after rampage 2 with three blockers", power, toughness)
	}
	for _, blocker := range blockers {
		if _, alive := permanentByObjectID(after, blocker.ObjectID); alive {
			t.Fatalf("blocker %v survived the 6/7 trampler", blocker.ObjectID)
		}
	}
	if got := after.Players[game.Player2].Life; got != life-3 {
		t.Fatalf("defender life = %d, want %d after 3 trample damage", got, life-3)
	}
}

// TestPhyrexianColossusWithMenaceNeedsThree: granting menace to Phyrexian
// Colossus doesn't add the minimums. It still needs three blockers, not two
// and not five.
func TestPhyrexianColossusWithMenaceNeedsThree(t *testing.T) {
	g, colossus, blockers := threeOrMoreAttack(exceptByThreeOrMore("Phyrexian Colossus", 8, 8, game.Menace), 5)
	if got := attackerMinimumBlockers(g, colossus); got != 3 {
		t.Fatalf("minimum blockers = %d, want 3 (the larger of menace and three, not their sum)", got)
	}
	checkBlockLegality(t, g, map[string]struct {
		blocks []game.BlockDeclaration
		legal  bool
	}{
		"two (menace alone)": {blocksBy(colossus, blockers[0], blockers[1]), false},
		"three":              {blocksBy(colossus, blockers[0], blockers[1], blockers[2]), true},
		"four":               {blocksBy(colossus, blockers[:4]...), true},
	})
}

// TestLureOnThreeOrMoreAttacker: a true lure ("all creatures able to block it
// do so") on a creature that needs three blockers. With two able blockers no
// legal block satisfies the lure, so not blocking is legal. With three, all
// three must block.
func TestLureOnThreeOrMoreAttacker(t *testing.T) {
	g, pathrazer, blockers := threeOrMoreAttack(exceptByThreeOrMore("Pathrazer of Ulamog", 9, 9), 2)
	addRuleEffectSource(g, game.Player1, game.RuleEffectMustBeBlockedByAllAble, pathrazer.ObjectID)
	checkBlockLegality(t, g, map[string]struct {
		blocks []game.BlockDeclaration
		legal  bool
	}{
		"no block": {nil, true},
		"both":     {blocksBy(pathrazer, blockers...), false},
	})

	g, pathrazer, blockers = threeOrMoreAttack(exceptByThreeOrMore("Pathrazer of Ulamog", 9, 9), 3)
	addRuleEffectSource(g, game.Player1, game.RuleEffectMustBeBlockedByAllAble, pathrazer.ObjectID)
	checkBlockLegality(t, g, map[string]struct {
		blocks []game.BlockDeclaration
		legal  bool
	}{
		"no block":  {nil, false},
		"two":       {blocksBy(pathrazer, blockers[0], blockers[1]), false},
		"all three": {blocksBy(pathrazer, blockers...), true},
	})
	for _, count := range offeredBlockCounts(t, g, pathrazer) {
		if count != 3 {
			t.Fatalf("offered a %d-creature block, want only the forced three", count)
		}
	}

	// "Must be blocked if able" alone needs three blockers to be obeyed.
	g, pathrazer, blockers = threeOrMoreAttack(exceptByThreeOrMore("Pathrazer of Ulamog", 9, 9), 4)
	addRuleEffectSource(g, game.Player1, game.RuleEffectMustBeBlocked, pathrazer.ObjectID)
	checkBlockLegality(t, g, map[string]struct {
		blocks []game.BlockDeclaration
		legal  bool
	}{
		"no block": {nil, false},
		"two":      {blocksBy(pathrazer, blockers[0], blockers[1]), false},
		"three":    {blocksBy(pathrazer, blockers[1], blockers[2], blockers[3]), true},
	})
}

// TestThreeOrMoreAndAtMostOneIsUnblockable: "can't be blocked by more than one
// creature" together with "except by three or more" leaves no legal block.
func TestThreeOrMoreAndAtMostOneIsUnblockable(t *testing.T) {
	g, guile, blockers := threeOrMoreAttack(exceptByThreeOrMore("Guile", 6, 6), 3)
	addRuleEffectSource(g, game.Player1, game.RuleEffectCantBeBlockedByMoreThanOne, guile.ObjectID)
	addRuleEffectSource(g, game.Player1, game.RuleEffectMustBeBlocked, guile.ObjectID)
	for _, count := range offeredBlockCounts(t, g, guile) {
		if count != 0 {
			t.Fatalf("offered a %d-creature block of an unblockable creature", count)
		}
	}
	checkBlockLegality(t, g, map[string]struct {
		blocks []game.BlockDeclaration
		legal  bool
	}{
		"no block": {nil, true},
		"one":      {blocksBy(guile, blockers[0]), false},
		"three":    {blocksBy(guile, blockers...), false},
	})
}
