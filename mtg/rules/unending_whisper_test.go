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
// Unending Whisper Oracle ID c5da5c31-0cf1-4393-8701-eda7880ecebc,
// printing fc48180a-ccac-469f-938d-c050821d0160 (Tarkir: Dragonstorm #62),
// {U}, Sorcery. Oracle text: "Draw a card. Harmonize {5}{U} (You may cast
// this card from your graveyard for its harmonize cost. You may tap a
// creature you control to reduce that cost by {X}, where X is its power.
// Then exile this spell.)"
// The Scryfall printing endpoint
// https://api.scryfall.com/cards/fc48180a-ccac-469f-938d-c050821d0160,
// retrieved 2026-11-21, matches the pinned text. Its rulings endpoint
// https://api.scryfall.com/cards/fc48180a-ccac-469f-938d-c050821d0160/rulings
// returned seven WotC rulings dated 2025-04-04: normal sorcery timing still
// applies; the tapped creature's power reduces only generic mana after cost
// increases; mana value remains based on the printed mana cost; and the card
// is exiled whenever its Harmonize spell leaves the stack. The rulings also
// confirm that casting is permitted for cards put into the graveyard without
// being cast and that a just-arrived card may be cast before another player
// acts when otherwise legal. Raw responses are retained under
// .cardwork/source/unending-whisper/.
// Read the official Wizards Comprehensive Rules effective November 6, 2026,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// verified local copy .cardwork/source/unending-whisper/MagicCompRules-20261106.txt.
// Rules 307.1 and 601.3 preserve sorcery timing despite the graveyard
// permission; 601.2b and 601.2f-h govern announcing the alternative cost,
// locking the total and paying it; 118.7a restricts the reduction to generic
// mana; 702.180a-b define the Harmonize permission, optional creature choice,
// reduction, tap during payment and exile replacement; 121.1 defines drawing
// as moving the top card of the library into its owner's hand.
func TestUnendingWhisperHarmonizeDrawsAndExiles(t *testing.T) {
	const oracleID = "c5da5c31-0cf1-4393-8701-eda7880ecebc"
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("lookup Unending Whisper by Oracle ID: %v", err)
	}
	if def.Name != "Unending Whisper" {
		t.Fatalf("Oracle ID resolved to %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToGraveyard(g, game.Player1, def)
	drawnID := addCardToLibrary(g, game.Player1, &game.CardDef{
		CardFace: game.CardFace{Name: "Drawn card"},
	})
	chosen := addCreatureWithPowerToughness(g, game.Player1, 5, 5)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player1)

	cast := action.CastSpellFromZone(spellID, zone.Graveyard, nil, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), cast) {
		t.Fatal("Unending Whisper Harmonize cast was not legal with {U} and the chosen creature's reduction")
	}
	if !engine.applyActionWithChoices(g, game.Player1, cast, harmonizeAgent(t, chosen, false), &TurnLog{}) {
		t.Fatal("could not cast Unending Whisper from the graveyard using Harmonize")
	}
	spell, ok := g.Stack.Peek()
	if !ok || spell.AlternativeCost != 1 || spell.HarmonizeCreature != opt.Val(chosen.ObjectID) ||
		spell.ManaSpentToCast != 1 || g.Players[game.Player1].ManaPool.Total() != 0 || !chosen.Tapped {
		t.Fatalf("Unending Whisper did not pay {U} and tap the chosen creature: %+v", spell)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Hand.Contains(drawnID) {
		t.Fatal("Unending Whisper did not draw the top card of its controller's library")
	}
	if !g.Players[game.Player1].Exile.Contains(spellID) ||
		g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("Unending Whisper cast with Harmonize was not exiled after resolving")
	}
}
