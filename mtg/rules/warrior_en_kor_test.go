package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Warrior en-Kor, Oracle 1245b165-e26f-4839-8468-2e95148fc6f6, printing
// afcf1e51-5e9a-4c86-8a14-97d9e919c320: {W}{W}, Creature — Kor Warrior
// Knight, 2/2; "{0}: The next 1 damage that would be dealt to this creature
// this turn is dealt to target creature you control instead."
// Current Scryfall printing and rulings were fetched 2026-11-21; the exact
// responses are retained in .cardwork/sources/en-kor/printing.json and
// rulings.json. Its five WotC rulings say redirected combat damage remains
// combat damage; a recipient may be dealt damage beyond its toughness; the
// ability may be activated repeatedly before damage; it does not stop trample
// damage assigned to the defending player; and it can redirect damage to
// itself.
// Official Comprehensive Rules effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// verified local copy SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829.
// Read §§115.1c, 602.1a, 602.2a-b, and 614.9: the target is chosen when the
// ability is activated, {0} is its activation cost, and damage is redirected
// to the chosen creature rather than prevented.
const warriorEnKorTestOracleID = "1245b165-e26f-4839-8468-2e95148fc6f6"

func TestWarriorEnKorRedirectsOneDamageToTargetCreature(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(warriorEnKorTestOracleID)
	if err != nil {
		t.Fatalf("exact Warrior en-Kor lookup failed: %v", err)
	}
	if def.Name != "Warrior en-Kor" {
		t.Fatalf("exact Warrior en-Kor lookup returned %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	warrior := addCombatPermanent(g, game.Player1, def)
	recipient := addCombatPermanent(g, game.Player1, redirectTestCard(
		"Warrior en-Kor's chosen creature", []types.Card{types.Creature}, 1, 5,
	))
	engine := NewEngine(nil)
	target := game.PermanentTarget(recipient.ObjectID)
	agents := [game.NumPlayers]PlayerAgent{}

	activateRedirect(t, engine, g, game.Player1, warrior, 0, []game.Target{target}, agents)
	resolveStackWithTriggers(engine, g, agents)
	boltOnStack(g, game.Player2, game.PermanentTarget(warrior.ObjectID))
	resolveStackWithTriggers(engine, g, agents)

	if warrior.MarkedDamage != 2 || recipient.MarkedDamage != 1 || g.Players[game.Player1].Life != 40 {
		t.Fatalf("after 3 damage to Warrior en-Kor: Warrior damage %d, chosen creature damage %d, controller life %d; want 2, 1, 40",
			warrior.MarkedDamage, recipient.MarkedDamage, g.Players[game.Player1].Life)
	}
}
