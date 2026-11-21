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
// Channeled Dragonfire Oracle ID 91179937-6f55-4d1b-8f06-2efe1c8fea7a,
// printing 24204881-690c-4043-8771-20cb93385072 (Tarkir: Dragonstorm #102),
// {R}, Sorcery. Oracle text: "Channeled Dragonfire deals 2 damage to any
// target. Harmonize {5}{R}{R} (You may cast this card from your graveyard for
// its harmonize cost. You may tap a creature you control to reduce that cost
// by {X}, where X is its power. Then exile this spell.)"
// The current Scryfall card endpoint
// https://api.scryfall.com/cards/24204881-690c-4043-8771-20cb93385072
// retrieved 2026-11-21 matches the pinned text; its rulings endpoint
// https://api.scryfall.com/cards/24204881-690c-4043-8771-20cb93385072/rulings
// returned seven WotC rulings dated 2025-04-04. They require ordinary
// sorcery timing, generic-only cost reduction based on the tapped creature's
// power after cost increases, unchanged mana value, and exile whenever the
// Harmonize spell leaves the stack. Raw responses are retained under
// .cardwork/source/channeled-dragonfire/.
// Read the official Wizards Comprehensive Rules effective November 6, 2026 at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (verified local copy .cardwork/source/channeled-dragonfire/MagicCompRules-20261106.txt):
// 115.1a and 115.4 govern target announcement and the creatures, players,
// planeswalkers, and battles allowed by "any target"; 120.1-3 govern damage
// and its result on a player. Rules 307.1 and 601.3 require normal sorcery
// timing and casting permission; 601.2a-i govern spell announcement, targets,
// locked total cost, payment, and casting. Rule 118.7a limits the reduction to
// generic mana, 608.2b rechecks targets on resolution, and 702.180a-b grant
// graveyard casting, reduce generic cost by the tapped creature's power, tap
// that creature during payment, and exile the spell when it leaves the stack.
func TestChanneledDragonfireHarmonizeDealsTwoDamage(t *testing.T) {
	const oracleID = "91179937-6f55-4d1b-8f06-2efe1c8fea7a"
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("lookup Channeled Dragonfire by Oracle ID: %v", err)
	}
	if def.Name != "Channeled Dragonfire" {
		t.Fatalf("Oracle ID resolved to %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToGraveyard(g, game.Player1, def)
	chosen := addCreatureWithPowerToughness(g, game.Player1, 5, 5)
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)
	setMainPhasePriority(g, game.Player1)
	startingLife := g.Players[game.Player2].Life

	cast := action.CastSpellFromZone(
		spellID, zone.Graveyard, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil,
	)
	if !containsAction(engine.legalActions(g, game.Player1), cast) {
		t.Fatal("Channeled Dragonfire Harmonize cast was not legal with the chosen creature's reduction")
	}
	if !engine.applyActionWithChoices(g, game.Player1, cast, harmonizeAgent(t, chosen, false), &TurnLog{}) {
		t.Fatal("could not cast Channeled Dragonfire from the graveyard using Harmonize")
	}
	spell, ok := g.Stack.Peek()
	if !ok || spell.AlternativeCost != 1 || spell.HarmonizeCreature != opt.Val(chosen.ObjectID) ||
		spell.ManaSpentToCast != 2 || g.Players[game.Player1].ManaPool.Total() != 0 || !chosen.Tapped {
		t.Fatalf("Channeled Dragonfire did not pay {R}{R} and tap the chosen creature: %+v", spell)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player2].Life; got != startingLife-2 {
		t.Fatalf("target player life = %d, want %d after 2 damage", got, startingLife-2)
	}
}
