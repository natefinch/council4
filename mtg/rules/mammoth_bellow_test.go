package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned-corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Mammoth Bellow Oracle ID 6f2e4607-bece-42ba-a2a3-5dc50f246793, printing
// 468b17b4-79ce-4dfa-8873-a9cfc347e38f (Tarkir: Dragonstorm #205),
// {2}{G}{U}{R}, Sorcery. Oracle text: "Create a 5/5 green Elephant creature
// token. Harmonize {5}{G}{U}{R} (You may cast this card from your graveyard for
// its harmonize cost. You may tap a creature you control to reduce that cost by
// {X}, where X is its power. Then exile this spell.)"
// The current Scryfall card endpoint
// https://api.scryfall.com/cards/468b17b4-79ce-4dfa-8873-a9cfc347e38f
// retrieved 2026-11-21 matches the pinned text; its rulings endpoint
// https://api.scryfall.com/cards/468b17b4-79ce-4dfa-8873-a9cfc347e38f/rulings
// returned seven WotC rulings dated 2025-04-04. They require ordinary sorcery
// timing, generic-only reduction from the tapped creature's power after cost
// increases, unchanged mana value, and exile whenever the Harmonize spell
// leaves the stack.
// Read the official Wizards Comprehensive Rules effective November 6, 2026 at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (verified local copy .cardwork/source/mammoth-bellow/MagicCompRules-20261106.txt):
// 111.1-4 establish token ownership/controller, defined characteristics and
// the Elephant Token name; 118.7a limits the reduction to generic mana; 307.1
// and 601.3 require normal sorcery timing and casting permission; 601.2a-i
// govern casting, alternative-cost announcement, locking and payment; 608.2c
// applies the spell instructions; 702.180a-b grant graveyard casting, reduce
// generic cost, exile the spell on stack departure, and require choosing the
// creature before tapping it during payment.
func TestMammothBellowHarmonizeCreatesElephant(t *testing.T) {
	const oracleID = "6f2e4607-bece-42ba-a2a3-5dc50f246793"
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("lookup Mammoth Bellow by Oracle ID: %v", err)
	}
	if def.Name != "Mammoth Bellow" {
		t.Fatalf("Oracle ID resolved to %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToGraveyard(g, game.Player1, def)
	chosen := addCreatureWithPowerToughness(g, game.Player1, 4, 4)
	for _, c := range []mana.Color{mana.C, mana.G, mana.U, mana.R} {
		g.Players[game.Player1].ManaPool.Add(c, 1)
	}
	setMainPhasePriority(g, game.Player1)

	cast := action.CastSpellFromZone(spellID, zone.Graveyard, nil, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), cast) {
		t.Fatal("Mammoth Bellow Harmonize cast was not legal with the chosen creature's reduction")
	}
	if !engine.applyActionWithChoices(g, game.Player1, cast, harmonizeAgent(t, chosen, false), &TurnLog{}) {
		t.Fatal("could not cast Mammoth Bellow from the graveyard using Harmonize")
	}
	spell, ok := g.Stack.Peek()
	if !ok || spell.AlternativeCost != 1 || spell.HarmonizeCreature != opt.Val(chosen.ObjectID) ||
		spell.ManaSpentToCast != 4 || g.Players[game.Player1].ManaPool.Total() != 0 || !chosen.Tapped {
		t.Fatalf("Mammoth Bellow did not pay {1}{G}{U}{R} and tap the chosen creature: %+v", spell)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Exile.Contains(spellID) ||
		g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("resolved Harmonize spell was not exiled")
	}

	var elephants []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.TokenDef != nil && permanent.TokenDef.Name == "Elephant Token" {
			elephants = append(elephants, permanent)
		}
	}
	if len(elephants) != 1 {
		t.Fatalf("created %d Elephant Tokens, want exactly one", len(elephants))
	}
	token := elephants[0]
	if token.Owner != game.Player1 || token.Controller != game.Player1 ||
		!slices.Contains(token.TokenDef.Types, types.Creature) ||
		!slices.Contains(token.TokenDef.Subtypes, types.Elephant) ||
		!slices.Contains(token.TokenDef.Colors, color.Green) ||
		!token.TokenDef.Power.Exists || token.TokenDef.Power.Val.Value != 5 ||
		!token.TokenDef.Toughness.Exists || token.TokenDef.Toughness.Val.Value != 5 {
		t.Fatalf("created token does not match the 5/5 green Elephant creature: %+v", token)
	}
}
