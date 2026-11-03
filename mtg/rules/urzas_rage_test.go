package rules

import (
	"testing"

	cardregistry "github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: pinned-corpus Oracle ID 363f8c66-fe0c-44b9-987d-1d160e3f9c54
// (registered by name; the legacy definition carries no Oracle ID). Scryfall
// rulings for that Oracle ID returned eight WotC rulings (2019-06-14,
// 2024-11-08), including that a copy of a kicked spell is also kicked.
// CR 2026-09-25 (verified official copy) 615.12 makes prevention effects prevent
// nothing from unpreventable damage and leaves shields unreduced; 702.33
// governs kicker and 707.10 spell copies.

// castUrzasRageThroughShield casts the registered Urza's Rage at Player2 under a
// global "prevent all damage" shield, optionally kicked and optionally copied,
// resolves the stack, and returns Player2's life lost.
func castUrzasRageThroughShield(t *testing.T, kicked, copied bool) int {
	t.Helper()
	registry := cardregistry.NewDefaultRegistry()
	def, err := registry.LookupChecked("Urza's Rage")
	if err != nil {
		t.Fatal(err)
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, registry.Programs())
	setMainPhasePriority(g, game.Player1)
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player3}, game.PreventDamage{All: true, Global: true}, nil)
	card := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)
	g.Players[game.Player1].ManaPool.Add(mana.C, 10)
	life := g.Players[game.Player2].Life
	targets := []game.Target{game.PlayerTarget(game.Player2)}

	cast := action.CastSpell(card, targets, 0, nil)
	if kicked {
		cast = action.CastKickedSpell(card, targets, 0, nil)
	}
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatalf("cast Urza's Rage kicked=%v failed", kicked)
	}
	original, ok := g.Stack.Peek()
	if !ok || original.KickerPaid != kicked {
		t.Fatalf("stack top = %+v, want Urza's Rage with KickerPaid=%v", original, kicked)
	}
	if copied {
		if copy := copyStackObject(engine, g, original); copy == nil || copy.ID == original.ID || copy.KickerPaid != kicked {
			t.Fatalf("copy = %+v, want a distinct copy with KickerPaid=%v", copy, kicked)
		}
		engine.resolveTopOfStack(g, &TurnLog{})
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Stack.Size() != 0 {
		t.Fatalf("stack has %d objects after resolution, want 0", g.Stack.Size())
	}
	if len(g.PreventionShields) != 1 {
		t.Fatalf("prevention shields = %+v, want the global shield intact", g.PreventionShields)
	}
	return life - g.Players[game.Player2].Life
}

// TestUrzasRageKickedDamageCantBePrevented proves the kicked Urza's Rage deals
// 10 unpreventable damage through a prevention shield, a kicked copy does the
// same, and the unkicked 3 damage is still prevented.
func TestUrzasRageKickedDamageCantBePrevented(t *testing.T) {
	if lost := castUrzasRageThroughShield(t, false, false); lost != 0 {
		t.Fatalf("unkicked damage through shield = %d, want prevented", lost)
	}
	if lost := castUrzasRageThroughShield(t, true, false); lost != 10 {
		t.Fatalf("kicked damage through shield = %d, want 10", lost)
	}
	if lost := castUrzasRageThroughShield(t, true, true); lost != 20 {
		t.Fatalf("kicked spell and kicked copy damage through shield = %d, want 20", lost)
	}
}
