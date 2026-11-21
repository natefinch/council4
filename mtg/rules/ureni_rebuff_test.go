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

// Source review: pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Ureni's Rebuff Oracle ID c0b7e989-8194-484c-93aa-a44df0764a77, printing
// 722716df-9cea-40a7-924b-c28497e227e6 (Tarkir: Dragonstorm #63), {1}{U},
// Sorcery. Exact Oracle text: "Return target creature to its owner's hand.
// Harmonize {5}{U} (You may cast this card from your graveyard for its
// harmonize cost. You may tap a creature you control to reduce that cost by
// {X}, where X is its power. Then exile this spell.)"
// The current Scryfall printing endpoint
// https://api.scryfall.com/cards/722716df-9cea-40a7-924b-c28497e227e6,
// retrieved 2026-11-21, matches the pinned printing. Its rulings endpoint
// https://api.scryfall.com/cards/722716df-9cea-40a7-924b-c28497e227e6/rulings
// returned an empty list.
// Read the official Magic: The Gathering Comprehensive Rules effective
// 2026-11-06 at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (verified local copy in the retained handoff,
// .cardwork/MagicCompRules-20261106.txt). Rules 115.1a and 601.2c require a
// legal creature target to be chosen while casting; 608.2b rechecks it on
// resolution. Rule 400.3 sends the creature to its owner's hand, and 307.1
// requires normal sorcery timing. Rules 601.2b/f-h and 118.7a govern announcing
// and paying the alternative cost and limit the reduction to generic mana.
// Rules 702.180a-b grant the graveyard cast, bind the generic reduction to the
// chosen creature's power, tap that creature during payment, and exile the
// spell when it leaves the stack.
func TestUrenisRebuffHarmonizeReturnsTargetCreatureToOwnersHand(t *testing.T) {
	const oracleID = "c0b7e989-8194-484c-93aa-a44df0764a77"
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("lookup Ureni's Rebuff by Oracle ID: %v", err)
	}
	if def.Name != "Ureni's Rebuff" {
		t.Fatalf("Oracle ID resolved to %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToGraveyard(g, game.Player1, def)
	chosen := harmonizeCreature(g, "Harmonize payment creature", 5)
	target := addCreaturePermanent(g, game.Player2)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player1)

	cast := action.CastSpellFromZone(
		spellID,
		zone.Graveyard,
		[]game.Target{game.PermanentTarget(target.ObjectID)},
		0,
		nil,
	)
	if !containsAction(engine.legalActions(g, game.Player1), cast) {
		t.Fatal("Ureni's Rebuff Harmonize cast was not legal with the chosen creature's reduction")
	}
	if !engine.applyActionWithChoices(g, game.Player1, cast, harmonizeAgent(t, chosen, false), &TurnLog{}) {
		t.Fatal("could not cast Ureni's Rebuff from the graveyard using Harmonize")
	}
	spell, ok := g.Stack.Peek()
	if !ok || spell.AlternativeCost != 1 || spell.HarmonizeCreature != opt.Val(chosen.ObjectID) ||
		spell.ManaSpentToCast != 1 || g.Players[game.Player1].ManaPool.Total() != 0 || !chosen.Tapped {
		t.Fatalf("Ureni's Rebuff did not pay {U} and tap the chosen 5-power creature: %+v", spell)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := permanentByObjectID(g, target.ObjectID); ok {
		t.Fatal("target creature remained on the battlefield")
	}
	if !g.Players[game.Player2].Hand.Contains(target.CardInstanceID) {
		t.Fatal("target creature was not returned to its owner's hand")
	}
}
