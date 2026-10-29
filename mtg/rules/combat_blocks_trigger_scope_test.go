package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

// Source review: official CR effective 2026-09-25 (verified copy
// MagicCompRules 20260925.txt), read 509.1g-i, 509.3a-d and 702.22h. 509.3a:
// "Whenever [a creature] blocks" triggers only once each combat for that
// creature, even if it blocks multiple creatures; 509.3b: "Whenever [a
// creature] blocks a creature" triggers once for each attacking creature it
// blocks. 702.22h: a creature blocking one member of a band blocks every member.
// Card faces (pinned corpus): Goldenglow Moth ({W} 0/1 Flying; "Whenever this
// creature blocks, you may gain 4 life."), Konda, Lord of Eiganjo ({5}{W}{W}
// 3/3 Vigilance, indestructible, Bushido 5 "Whenever this creature blocks or
// becomes blocked, it gets +5/+5 until end of turn."), Carnage Gladiator
// ({2}{B}{R} 4/2 "Whenever a creature blocks, that creature's controller loses
// 1 life."; Scryfall ruling 2017-03-14: if a creature can block multiple
// creatures and does so, the ability triggers only once), Netcaster Spider
// ({2}{G} 2/3 Reach; "Whenever this creature blocks a creature with flying,
// this creature gets +2/+0 until end of turn."), Shield of the Righteous
// ({W}{U} Equipment; "Equipped creature gets +0/+2 and has vigilance. Whenever
// equipped creature blocks a creature, that creature doesn't untap during its
// controller's next untap step." Equip {2}), Benalish Hero ({W} 1/1 Banding),
// Mesa Pegasus ({1}{W} 1/1 Flying, banding), Kjeldoran Escort ({2}{W}{W} 2/3
// Banding). Retained responses: .cardwork/blocks_scope_oracle.json and
// .cardwork/rulings_*.json in the unit worktree.

// TestBlocksTriggerFiresOnceWhenBlockingBand runs registered Goldenglow Moth
// blocking a two-member band: it blocks both members (CR 702.22h) but its
// "blocks" trigger fires once (CR 509.3a).
func TestBlocksTriggerFiresOnceWhenBlockingBand(t *testing.T) {
	g := bandCombatGame()
	hero := registryCombatPermanent(t, g, game.Player1, "Benalish Hero")
	bear := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	moth := registryCombatPermanent(t, g, game.Player2, "Goldenglow Moth")
	defender := &bandBlockingAgent{blocks: []game.BlockDeclaration{{Blocker: moth.ObjectID, Blocking: hero.ObjectID}}}
	policies := simPassPolicies()
	policies[game.Player1], policies[game.Player2] = &bandingDivisionAgent{}, defender
	life := g.Players[game.Player2].Life

	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, bandAttack(hero, bear), policies)
	if !ok {
		t.Fatal("combat not accepted")
	}
	if got := len(blockEvents(after.Events, game.EventBlockerDeclared)); got != 2 {
		t.Fatalf("block pairs = %d, want 2 (Goldenglow Moth blocks both band members)", got)
	}
	if got := after.Players[game.Player2].Life; got != life+4 {
		t.Fatalf("defender life = %d, want %d: Goldenglow Moth's blocks trigger fires once (CR 509.3a)", got, life+4)
	}
}

// TestBlocksTriggerFiresOnceWhenBlockingTwoAttackers runs registered Goldenglow
// Moth, given "can block an additional creature each combat", blocking two
// separate attackers: its "blocks" trigger still fires once (CR 509.3a).
func TestBlocksTriggerFiresOnceWhenBlockingTwoAttackers(t *testing.T) {
	g := bandCombatGame()
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	def := registryCardDef(t, "Goldenglow Moth")
	def.StaticAbilities = append(slices.Clone(def.StaticAbilities), game.StaticAbility{
		RuleEffects: []game.RuleEffect{{
			Kind:                 game.RuleEffectCanBlockAdditional,
			AffectedSource:       true,
			AdditionalBlockCount: 1,
		}},
	})
	moth := addCombatPermanent(g, game.Player2, def)
	defender := &bandBlockingAgent{blocks: []game.BlockDeclaration{
		{Blocker: moth.ObjectID, Blocking: first.ObjectID},
		{Blocker: moth.ObjectID, Blocking: second.ObjectID},
	}}
	policies := simPassPolicies()
	policies[game.Player2] = defender
	life := g.Players[game.Player2].Life
	attack := action.DeclareAttackersAction{}
	for _, attacker := range []*game.Permanent{first, second} {
		attacker.SummoningSick = false
		attack.Attackers = append(attack.Attackers, attackAt(attacker, game.AttackTarget{Player: game.Player2}))
	}

	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, attack, policies)
	if !ok {
		t.Fatal("combat not accepted")
	}
	if got := len(blockEvents(after.Events, game.EventBlockerDeclared)); got != 2 {
		t.Fatalf("block pairs = %d, want 2 (Goldenglow Moth blocks both attackers)", got)
	}
	if got := after.Players[game.Player2].Life; got != life+4 {
		t.Fatalf("defender life = %d, want %d: Goldenglow Moth's blocks trigger fires once (CR 509.3a)", got, life+4)
	}
}

// TestBlocksOrBecomesBlockedTriggerFiresOnceWhenBlockingBand runs registered
// Konda, Lord of Eiganjo blocking a two-member band: the "blocks" half of
// bushido 5 fires once, so Konda is 8/8, not 13/13 (CR 509.3a).
func TestBlocksOrBecomesBlockedTriggerFiresOnceWhenBlockingBand(t *testing.T) {
	g := bandCombatGame()
	hero := registryCombatPermanent(t, g, game.Player1, "Benalish Hero")
	bear := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	konda := registryCombatPermanent(t, g, game.Player2, "Konda, Lord of Eiganjo")
	defender := &bandBlockingAgent{blocks: []game.BlockDeclaration{{Blocker: konda.ObjectID, Blocking: hero.ObjectID}}}
	policies := simPassPolicies()
	policies[game.Player1], policies[game.Player2] = &bandingDivisionAgent{}, defender

	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, bandAttack(hero, bear), policies)
	if !ok {
		t.Fatal("combat not accepted")
	}
	afterKonda, ok := permanentByObjectID(after, konda.ObjectID)
	if !ok {
		t.Fatal("Konda left the battlefield")
	}
	if got := effectivePower(after, afterKonda); got != 8 {
		t.Fatalf("Konda power = %d, want 8: bushido's blocks trigger fires once (CR 509.3a)", got)
	}
}

// TestCreatureBlocksTriggerFiresOncePerBlocker runs registered Carnage
// Gladiator while two creatures each block a two-member band, so each blocks
// two attackers: the defender loses 1 life per blocking creature, 2 in all
// (CR 509.3a; Carnage Gladiator ruling 2017-03-14).
func TestCreatureBlocksTriggerFiresOncePerBlocker(t *testing.T) {
	g := bandCombatGame()
	registryCombatPermanent(t, g, game.Player1, "Carnage Gladiator")
	hero := registryCombatPermanent(t, g, game.Player1, "Benalish Hero")
	bear := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	defender := &bandBlockingAgent{blocks: []game.BlockDeclaration{
		{Blocker: first.ObjectID, Blocking: hero.ObjectID},
		{Blocker: second.ObjectID, Blocking: bear.ObjectID},
	}}
	policies := simPassPolicies()
	policies[game.Player1], policies[game.Player2] = &bandingDivisionAgent{}, defender
	life := g.Players[game.Player2].Life

	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, bandAttack(hero, bear), policies)
	if !ok {
		t.Fatal("combat not accepted")
	}
	if got := len(blockEvents(after.Events, game.EventBlockerDeclared)); got != 4 {
		t.Fatalf("block pairs = %d, want 4 (each blocker blocks both band members)", got)
	}
	if got := after.Players[game.Player2].Life; got != life-2 {
		t.Fatalf("defender life = %d, want %d: Carnage Gladiator triggers once per blocking creature", got, life-2)
	}
}

// TestBlocksACreatureTriggerFiresPerBlockedFlyer runs registered Netcaster
// Spider blocking a band of two flyers: "blocks a creature with flying"
// triggers once for each (CR 509.3b), so the Spider is 6/3.
func TestBlocksACreatureTriggerFiresPerBlockedFlyer(t *testing.T) {
	g := bandCombatGame()
	pegasus := registryCombatPermanent(t, g, game.Player1, "Mesa Pegasus")
	flyer := addCombatCreaturePermanentWithPower(g, game.Player1, 1, game.Flying)
	spider := registryCombatPermanent(t, g, game.Player2, "Netcaster Spider")
	defender := &bandBlockingAgent{blocks: []game.BlockDeclaration{{Blocker: spider.ObjectID, Blocking: pegasus.ObjectID}}}
	policies := simPassPolicies()
	policies[game.Player1], policies[game.Player2] = &bandingDivisionAgent{}, defender

	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, bandAttack(pegasus, flyer), policies)
	if !ok {
		t.Fatal("combat not accepted")
	}
	afterSpider, ok := permanentByObjectID(after, spider.ObjectID)
	if !ok {
		t.Fatal("Netcaster Spider left the battlefield")
	}
	if got := effectivePower(after, afterSpider); got != 6 {
		t.Fatalf("Netcaster Spider power = %d, want 6: blocks-a-creature triggers per blocked flyer (CR 509.3b)", got)
	}
}

// TestEquippedBlocksACreatureTriggerFiresPerBandMember runs registered Shield of
// the Righteous on a creature blocking a two-member band: its "blocks a
// creature" trigger fires for each blocked member (CR 509.3b), so neither
// attacker untaps during its controller's next untap step.
func TestEquippedBlocksACreatureTriggerFiresPerBandMember(t *testing.T) {
	g := bandCombatGame()
	escort := registryCombatPermanent(t, g, game.Player1, "Kjeldoran Escort")
	bear := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
	shield := registryCombatPermanent(t, g, game.Player2, "Shield of the Righteous")
	attachPlate(blocker, shield)
	defender := &bandBlockingAgent{blocks: []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: escort.ObjectID}}}
	policies := simPassPolicies()
	policies[game.Player1], policies[game.Player2] = &bandingDivisionAgent{}, defender

	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, bandAttack(escort, bear), policies)
	if !ok {
		t.Fatal("combat not accepted")
	}
	for _, attacker := range []*game.Permanent{escort, bear} {
		afterAttacker, ok := permanentByObjectID(after, attacker.ObjectID)
		if !ok {
			t.Fatalf("attacker %v left the battlefield", attacker.ObjectID)
		}
		if !afterAttacker.Exerted {
			t.Fatalf("attacker %v will untap: Shield of the Righteous triggers per blocked creature (CR 509.3b)", attacker.ObjectID)
		}
	}
}
