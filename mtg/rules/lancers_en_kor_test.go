package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Lancers en-Kor, Oracle ef49dc78-9fd9-4cf8-af10-5a6b8ef7fc57, printing
// e7cd99c2-6d4a-48e8-8848-6fdc0788525d: {3}{W}{W}, Creature — Kor Soldier,
// 3/3; "Trample / {0}: The next 1 damage that would be dealt to this creature
// this turn is dealt to target creature you control instead." The pinned
// record and Scryfall printing fetched 2026-11-21 agree; the exact response is
// retained at .cardwork/sources/en-kor/printing-e7cd99c2-6d4a-48e8-8848-6fdc0788525d.json.
// The 2026-11-21 Scryfall rulings response (HTTP 200, five WotC rulings) at
// https://api.scryfall.com/cards/e7cd99c2-6d4a-48e8-8848-6fdc0788525d/rulings
// confirms redirected combat damage remains combat damage; the ability does
// not stop trample damage from being assigned to the defending player; it can
// redirect damage to itself; a recipient can be dealt more than its toughness;
// and the ability may be activated repeatedly before damage. The exact
// response is retained at
// .cardwork/sources/en-kor/rulings-e7cd99c2-6d4a-48e8-8848-6fdc0788525d.json.
// Official Comprehensive Rules effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// verified local copy .cardwork/sources/en-kor/MagicCompRules-20261106-live.txt,
// SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829.
// Read §§115.1c, 601.2c, 602.1a, 602.2a-b: activation pays {0} and chooses
// the target creature you control. §§510.1a-d and 702.19a-b require an
// attacking trample creature to assign lethal damage to its blocker before
// assigning excess to the defending player. §§120.1, 120.2a, 120.4b,
// 614.1a, 614.5, 614.9 and 616.1 require the next one damage to be redirected
// as the same damage, after assignment, not prevented; its source and combat
// status remain relevant, and other applicable replacements are ordered by
// the affected player.
func TestLancersEnKorTramplesAndRedirectsOneCombatDamage(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID("ef49dc78-9fd9-4cf8-af10-5a6b8ef7fc57")
	if err != nil {
		t.Fatalf("exact Lancers en-Kor lookup failed: %v", err)
	}
	if def.Name != "Lancers en-Kor" {
		t.Fatalf("exact Lancers en-Kor lookup returned %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	lancers := addCombatPermanent(g, game.Player1, def)
	recipient := addCombatPermanent(g, game.Player1, redirectTestCard(
		"Lancers en-Kor's chosen creature", []types.Card{types.Creature}, 1, 5,
	))
	blocker := addCombatPermanent(g, game.Player2, redirectTestCard(
		"Blocking creature", []types.Card{types.Creature}, 2, 2,
	))
	engine := NewEngine(nil)
	target := game.PermanentTarget(recipient.ObjectID)
	agents := [game.NumPlayers]PlayerAgent{}

	activateRedirect(t, engine, g, game.Player1, lancers, 0, []game.Target{target}, agents)
	resolveStackWithTriggers(engine, g, agents)

	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepCombatDamage
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: lancers.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		Blockers:  []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: lancers.ObjectID}},
	}
	engine.resolveCombatDamage(g, &TurnLog{})

	if g.Players[game.Player2].Life != 39 || blocker.MarkedDamage != 2 ||
		lancers.MarkedDamage != 1 || recipient.MarkedDamage != 1 || len(g.PreventionShields) != 0 {
		t.Fatalf("after combat: defending life %d, blocker damage %d, Lancers damage %d, chosen creature damage %d, shields %+v; want 39, 2, 1, 1, none",
			g.Players[game.Player2].Life, blocker.MarkedDamage, lancers.MarkedDamage, recipient.MarkedDamage, g.PreventionShields)
	}
}
