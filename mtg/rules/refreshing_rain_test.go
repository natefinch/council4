package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: Refreshing Rain, Oracle ID c6f69168-71c2-41a2-ab95-cfd7baf0ba68,
// Nemesis #110, printing revision c5e24850-bd9b-40ab-878f-b8a554da1956;
// pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Current Oracle verified 2026-11-19:
// https://api.scryfall.com/cards/c5e24850-bd9b-40ab-878f-b8a554da1956
// {3}{G} Instant. "If an opponent controls a Swamp and you control a Forest,
// you may cast this spell without paying its mana cost. Target player gains 6 life."
// Rulings retrieved 2026-11-19 from
// https://api.scryfall.com/cards/c5e24850-bd9b-40ab-878f-b8a554da1956/rulings:
// the endpoint returned an empty list.
// Wizards Comprehensive Rules effective 2026-11-06:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// Read §§109.5, 115.1a, 118.9a-d, 119.3, 202.3, 305.6, 601.2b/c/f-h, 608.2b.
// "You" is the would-be caster; the Forest and any opponent's Swamp enable one
// optional alternate cost at casting without changing mana value. The target
// player is chosen as the spell is cast and gains 6 life on resolution.
const refreshingRainOracleID = "c6f69168-71c2-41a2-ab95-cfd7baf0ba68"

func TestAuthoredRefreshingRainScenario(t *testing.T) {
	def, err := cards.NewDefaultRegistry().LookupOracleID(refreshingRainOracleID)
	if err != nil {
		t.Fatalf("exact Refreshing Rain default-registry lookup failed: %v", err)
	}
	if def.Name != "Refreshing Rain" {
		t.Fatalf("exact Refreshing Rain lookup returned %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, def)
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	forest.Tapped = true
	swamp := addBasicLandPermanent(g, game.Player3, types.Swamp)
	swamp.Tapped = true
	setMainPhasePriority(g, game.Player1)

	targetLife := g.Players[game.Player2].Life
	casterLife := g.Players[game.Player1].Life
	cast := action.CastSpell(spellID, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), cast) {
		t.Fatal("Refreshing Rain was not available to cast for free with a Forest and an opponent-controlled Swamp")
	}
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("Refreshing Rain could not be cast using its conditional alternative cost")
	}
	if spell, ok := g.Stack.Peek(); !ok || spell.ManaSpentToCast != 0 {
		t.Fatalf("Refreshing Rain did not use its free alternative cost: %+v", spell)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if got, want := g.Players[game.Player2].Life, targetLife+6; got != want {
		t.Fatalf("target player's life = %d, want %d", got, want)
	}
	if got := g.Players[game.Player1].Life; got != casterLife {
		t.Fatalf("caster's life = %d, want unchanged at %d", got, casterLife)
	}
	if !g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("resolved Refreshing Rain did not reach its owner's graveyard")
	}
}
