package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: TDM 154, Oracle ID cf2355f3-11d2-4b03-b663-9fad6b07e4b2,
// pinned printing fb8c2d5c-ba0c-4d50-8898-5c6574b1e974, corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. The
// Oracle text is “Search your library for a basic land card, put it onto the
// battlefield tapped, then shuffle. Harmonize {4}{G} (You may cast this card
// from your graveyard for its harmonize cost. You may tap a creature you
// control to reduce that cost by {X}, where X is its power. Then exile this
// spell.)” Current Scryfall printing and rulings were retrieved 2026-11-21
// from https://api.scryfall.com/cards/fb8c2d5c-ba0c-4d50-8898-5c6574b1e974
// and https://api.scryfall.com/cards/fb8c2d5c-ba0c-4d50-8898-5c6574b1e974/rulings;
// the printing matches the pin and the HTTP 200 rulings response was empty.
// The saved batch rules pin is 2026-09-25; the official CR copy read here is
// effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829):
// 601.2b/f-h, 118.7a, and 702.180a-b require announcing the harmonize
// alternative and creature choice, locking a generic-only power reduction,
// then tapping that creature during payment. 110.5b requires the land to enter
// tapped as stated; 701.23a-b governs the hidden-zone search and optional
// find, 701.24a requires shuffling the library, and 701.26a permits tapping
// only an untapped permanent.
func TestRoamersRoutineHarmonizeFindsBasicLandTapped(t *testing.T) {
	const oracleID = "cf2355f3-11d2-4b03-b663-9fad6b07e4b2"
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatal(err)
	}
	if def.Name != "Roamer's Routine" {
		t.Fatalf("Oracle ID %s returned %q", oracleID, def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	spellID := addCardToGraveyard(g, game.Player1, def)
	creature := harmonizeCreature(g, "Power Two Creature", 2)
	basicID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:       "Basic Forest",
		Types:      []types.Card{types.Land},
		Supertypes: []types.Super{types.Basic},
		Subtypes:   []types.Sub{types.Forest},
	}})
	nonbasicID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Nonbasic Forest",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Forest},
	}})
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)

	cast := action.CastSpellFromZone(spellID, zone.Graveyard, nil, 0, nil)
	if !engine.applyActionWithChoices(
		g, game.Player1, cast, harmonizeAgent(t, creature, false), &TurnLog{},
	) {
		t.Fatal("Roamer's Routine could not be cast from the graveyard for its Harmonize cost")
	}
	if !creature.Tapped {
		t.Fatal("Roamer's Routine did not tap the chosen creature as part of paying for Harmonize")
	}
	spell, ok := g.Stack.Peek()
	if !ok || spell.AlternativeCost != 1 || !spell.HarmonizeCreature.Exists ||
		spell.HarmonizeCreature.Val != creature.ObjectID || spell.ManaSpentToCast != 3 {
		t.Fatalf("Roamer's Routine did not use the chosen creature's Harmonize cost: %+v", spell)
	}

	searchAgent := &suspendFamilySearchAgent{}
	engine.resolveTopOfStackWithChoices(g, suspendAgents(searchAgent), &TurnLog{})
	if len(searchAgent.requests) != 1 || len(searchAgent.requests[0].Options) != 1 {
		t.Fatalf("Roamer's Routine search choices = %+v, want only the basic land", searchAgent.requests)
	}
	land := suspendFamilyPermanent(t, g, basicID)
	if land.Controller != game.Player1 || !land.Tapped {
		t.Fatalf("found basic land controller/tapped = %v/%v, want Player1/true", land.Controller, land.Tapped)
	}
	if !g.Players[game.Player1].Library.Contains(nonbasicID) ||
		g.Players[game.Player1].Library.Contains(basicID) {
		t.Fatal("Roamer's Routine moved a nonbasic land or left the found basic land in the library")
	}
	if !g.Players[game.Player1].Exile.Contains(spellID) ||
		g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("Roamer's Routine did not exile the spell after resolving its Harmonize cast")
	}
}
