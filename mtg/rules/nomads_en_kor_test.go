package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Nomads en-Kor, Oracle b8d395a3-0bfe-45c3-bb3a-820d4f235b88, printing
// 48ed076a-909f-4ef0-b85e-a1935b1c2827: {W}, Creature — Kor Nomad Soldier,
// 1/1; "{0}: The next 1 damage that would be dealt to this creature this turn
// is dealt to target creature you control instead." The pinned record and
// current Scryfall printing fetched 2026-11-21 agree on the exact identity,
// characteristics, and Oracle text.
// The exact rulings response was fetched 2026-11-21 from
// https://api.scryfall.com/cards/48ed076a-909f-4ef0-b85e-a1935b1c2827/rulings
// (HTTP 200; five WotC rulings): redirected combat damage remains combat
// damage; the ability does not stop trample damage from being assigned to the
// defending player; it can redirect damage to itself; the recipient may take
// more damage than its toughness; and the ability may be used repeatedly
// before damage.
// The raw current Scryfall responses are retained under
// .cardwork/sources/en-kor/printing-48ed076a-909f-4ef0-b85e-a1935b1c2827.json
// and rulings-48ed076a-909f-4ef0-b85e-a1935b1c2827.json.
// Official Comprehensive Rules effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// verified local copy
// .cardwork/sources/en-kor/MagicCompRules-20261106.txt, SHA-256
// 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829.
// Read §§115.1c, 601.2c, 602.1a, 602.2a-b, 120.1, 120.3f, 614.1a, 614.5,
// 614.9, and 616.1: the target is chosen when the ability is activated; {0}
// is its activation cost; "instead" denotes a replacement effect; damage is
// redirected as the same damage to another legal recipient, with replacement
// effects applying in their rules-defined order. The original source and
// damage attributes remain relevant to the redirected damage.
func TestNomadsEnKorRedirectsOneDamageToTargetCreature(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID("b8d395a3-0bfe-45c3-bb3a-820d4f235b88")
	if err != nil {
		t.Fatalf("exact Nomads en-Kor lookup failed: %v", err)
	}
	if def.Name != "Nomads en-Kor" {
		t.Fatalf("exact Nomads en-Kor lookup returned %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	nomads := addCombatPermanent(g, game.Player1, def)
	recipient := addCombatPermanent(g, game.Player1, redirectTestCard(
		"Nomads en-Kor's chosen creature", []types.Card{types.Creature}, 1, 5,
	))
	engine := NewEngine(nil)
	target := game.PermanentTarget(recipient.ObjectID)
	agents := [game.NumPlayers]PlayerAgent{}

	activateRedirect(t, engine, g, game.Player1, nomads, 0, []game.Target{target}, agents)
	resolveStackWithTriggers(engine, g, agents)
	addEffectSpellToStack(g, game.Player2, game.Damage{
		Amount:    game.Fixed(1),
		Recipient: game.AnyTargetDamageRecipient(0),
	}, []game.Target{game.PermanentTarget(nomads.ObjectID)})
	resolveStackWithTriggers(engine, g, agents)

	if nomads.MarkedDamage != 0 || recipient.MarkedDamage != 1 || len(g.PreventionShields) != 0 {
		t.Fatalf("after one damage to Nomads en-Kor: Nomads damage %d, chosen creature damage %d, shields %+v; want 0, 1, none",
			nomads.MarkedDamage, recipient.MarkedDamage, g.PreventionShields)
	}
}
