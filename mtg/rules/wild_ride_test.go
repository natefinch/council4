package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned-corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Wild Ride Oracle ID 61978202-538b-4f92-acfc-e95f4a6ce466, printing
// abc8c6f5-6135-428e-8476-1751f82623f9 (Tarkir: Dragonstorm #132), {R},
// Sorcery. Oracle text: "Target creature gets +3/+0 and gains haste until end
// of turn. Harmonize {4}{R} (You may cast this card from your graveyard for
// its harmonize cost. You may tap a creature you control to reduce that cost
// by {X}, where X is its power. Then exile this spell.)"
// The current Scryfall card endpoint
// https://api.scryfall.com/cards/abc8c6f5-6135-428e-8476-1751f82623f9
// retrieved 2026-11-21 matches the pinned text; its rulings endpoint
// https://api.scryfall.com/cards/abc8c6f5-6135-428e-8476-1751f82623f9/rulings
// returned seven WotC rulings dated 2025-04-04. They require normal timing,
// reduce only generic mana by the chosen creature's power, apply other cost
// increases/reductions before locking the total, and exile the spell on every
// stack departure.
// Read the official Wizards Comprehensive Rules effective November 6, 2026 at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (verified local copy .cardwork/source/wild-ride/MagicCompRules-20261106.txt):
// 115.1a and 601.2c require a creature target; 601.2a-i, 601.3, 307.1,
// 118.7a and 702.180a-b govern announcement, ordinary sorcery timing,
// generic-only Harmonize reduction, and tapping the chosen creature while
// paying the locked total. Rule 608.2b rechecks the target; 611.2a/c and
// 613.1f-g govern the end-of-turn power and ability effects; 702.10b explains
// the granted haste.
func TestWildRideHarmonizeTargetsCreatureAndGrantsHaste(t *testing.T) {
	const oracleID = "61978202-538b-4f92-acfc-e95f4a6ce466"
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("lookup Wild Ride by Oracle ID: %v", err)
	}
	if def.Name != "Wild Ride" {
		t.Fatalf("Oracle ID resolved to %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToGraveyard(g, game.Player1, def)
	chosen := addCreatureWithPowerToughness(g, game.Player1, 3, 4)
	target := addCreatureWithPowerToughness(g, game.Player2, 2, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	setMainPhasePriority(g, game.Player1)

	cast := action.CastSpellFromZone(
		spellID, zone.Graveyard, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil,
	)
	if !engine.applyActionWithChoices(g, game.Player1, cast, harmonizeAgent(t, chosen, false), &TurnLog{}) {
		t.Fatal("could not cast Wild Ride from the graveyard using Harmonize")
	}
	spell, ok := g.Stack.Peek()
	if !ok || spell.AlternativeCost != 1 || spell.HarmonizeCreature != opt.Val(chosen.ObjectID) ||
		spell.ManaSpentToCast != 2 || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("Wild Ride did not pay {1}{R} after choosing the 3-power creature: %+v", spell)
	}
	if !chosen.Tapped {
		t.Fatal("Harmonize did not tap its chosen creature during payment")
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if got := effectivePower(g, target); got != 5 {
		t.Fatalf("target power = %d, want 5 after +3/+0", got)
	}
	if !hasKeyword(g, target, game.Haste) {
		t.Fatal("target did not gain haste until end of turn")
	}
	if !g.Players[game.Player1].Exile.Contains(spellID) ||
		g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("resolved Harmonize spell was not exiled")
	}
}
