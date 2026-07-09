package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestPreventNextSourceShieldPreventsMatchingColorOnce proves the one-shot
// color-filtered shield ("The next time a white source of your choice would
// deal damage to you this turn, prevent that damage.") prevents all of the
// next damage a matching-color source would deal to its controller and then
// expires, so a second such damage is unaffected.
func TestPreventNextSourceShieldPreventsMatchingColorOnce(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player2].Life = 20
	sourceID := addColoredSourceCard(g, game.Player1, color.White)
	obj := &game.StackObject{Controller: game.Player2}

	resolveInstruction(engine, g, obj, game.PreventDamage{
		Player:       game.ControllerReference(),
		SourceColors: []color.Color{color.White},
		All:          true,
		OneShot:      true,
	}, nil)

	if dealt := dealPlayerDamage(g, sourceID, 0, game.Player1, game.Player2, 4, false); dealt != 0 {
		t.Fatalf("first dealt damage = %d, want 0 after one-shot shield", dealt)
	}
	if g.Players[game.Player2].Life != 20 {
		t.Fatalf("player life = %d, want 20 after full prevention", g.Players[game.Player2].Life)
	}
	if len(g.PreventionShields) != 0 {
		t.Fatalf("prevention shields = %+v, want expired after one event", g.PreventionShields)
	}
	if dealt := dealPlayerDamage(g, sourceID, 0, game.Player1, game.Player2, 4, false); dealt != 4 {
		t.Fatalf("second dealt damage = %d, want 4 after shield expired", dealt)
	}
	if g.Players[game.Player2].Life != 16 {
		t.Fatalf("player life = %d, want 16", g.Players[game.Player2].Life)
	}
	assertEvent(t, g.Events, game.EventDamagePrevented, func(event game.Event) bool {
		return event.Player == game.Player2 && event.Amount == 4
	})
}

// TestPreventNextSourceShieldIgnoresOtherColor proves the color filter restricts
// the shield: damage from a source whose color does not match passes through
// undiminished and leaves the shield in place.
func TestPreventNextSourceShieldIgnoresOtherColor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player2].Life = 20
	redSource := addColoredSourceCard(g, game.Player1, color.Red)
	obj := &game.StackObject{Controller: game.Player2}

	resolveInstruction(engine, g, obj, game.PreventDamage{
		Player:       game.ControllerReference(),
		SourceColors: []color.Color{color.White},
		All:          true,
		OneShot:      true,
	}, nil)

	if dealt := dealPlayerDamage(g, redSource, 0, game.Player1, game.Player2, 4, false); dealt != 4 {
		t.Fatalf("dealt damage = %d, want 4 from a non-matching color source", dealt)
	}
	if len(g.PreventionShields) != 1 {
		t.Fatalf("prevention shields = %+v, want still present", g.PreventionShields)
	}
}

// TestPreventNextSourceShieldAnyColorPreventsOnce proves the unfiltered form
// ("a source of your choice") prevents the next damage from any source and then
// expires.
func TestPreventNextSourceShieldAnyColorPreventsOnce(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player2].Life = 20
	redSource := addColoredSourceCard(g, game.Player1, color.Red)
	obj := &game.StackObject{Controller: game.Player2}

	resolveInstruction(engine, g, obj, game.PreventDamage{
		Player:  game.ControllerReference(),
		All:     true,
		OneShot: true,
	}, nil)

	if dealt := dealPlayerDamage(g, redSource, 0, game.Player1, game.Player2, 5, false); dealt != 0 {
		t.Fatalf("dealt damage = %d, want 0 from any source", dealt)
	}
	if len(g.PreventionShields) != 0 {
		t.Fatalf("prevention shields = %+v, want expired", g.PreventionShields)
	}
}

// addPreventionSourceCard adds a graveyard-agnostic card instance (a resolved
// instant such as Deflecting Palm) that can act as a damage source, returning
// its CardInstance ID.
func addPreventionSourceCard(g *game.Game, owner game.PlayerID, name string) id.ID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:   name,
			Types:  []types.Card{types.Instant},
			Colors: []color.Color{color.Red, color.White},
		}},
		Owner: owner,
	}
	return cardID
}

// TestPreventNextSourceRedirectDealsPreventedAmountToSourceController proves the
// Deflecting Palm shield ("The next time a source of your choice would deal
// damage to you this turn, prevent that damage. If damage is prevented this way,
// Deflecting Palm deals that much damage to that source's controller.") prevents
// the next damage to its controller and deals that prevented amount back to the
// controller of the prevented source, as damage from the shield's own source.
func TestPreventNextSourceRedirectDealsPreventedAmountToSourceController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].Life = 20
	g.Players[game.Player2].Life = 20
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)
	palmID := addPreventionSourceCard(g, game.Player2, "Deflecting Palm")
	obj := &game.StackObject{Controller: game.Player2, SourceID: palmID}

	resolveInstruction(engine, g, obj, game.PreventDamage{
		Player:                              game.ControllerReference(),
		All:                                 true,
		OneShot:                             true,
		RedirectPreventedToSourceController: true,
	}, nil)

	if dealt := dealPlayerDamage(g, sourceID, 0, game.Player1, game.Player2, 4, false); dealt != 0 {
		t.Fatalf("dealt damage to shielded player = %d, want 0 after prevention", dealt)
	}
	if g.Players[game.Player2].Life != 20 {
		t.Fatalf("shielded player life = %d, want 20 after full prevention", g.Players[game.Player2].Life)
	}
	if g.Players[game.Player1].Life != 16 {
		t.Fatalf("source controller life = %d, want 16 after redirect", g.Players[game.Player1].Life)
	}
	if len(g.PreventionShields) != 0 {
		t.Fatalf("prevention shields = %+v, want expired after one event", g.PreventionShields)
	}
	assertEvent(t, g.Events, game.EventDamagePrevented, func(event game.Event) bool {
		return event.Player == game.Player2 && event.Amount == 4
	})
	assertEvent(t, g.Events, game.EventDamageDealt, func(event game.Event) bool {
		return event.Player == game.Player1 && event.SourceID == palmID && event.Amount == 4
	})
}

// TestPreventNextSourceRedirectIsItselfPreventable proves the redirected damage
// is a new damage event that the source's controller can prevent with their own
// one-shot shield, so composed Deflecting Palms cancel out.
func TestPreventNextSourceRedirectIsItselfPreventable(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player1].Life = 20
	g.Players[game.Player2].Life = 20
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)
	palmID := addPreventionSourceCard(g, game.Player2, "Deflecting Palm")

	// Player1 pre-emptively shields against the redirect.
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player1}, game.PreventDamage{
		Player:  game.ControllerReference(),
		All:     true,
		OneShot: true,
	}, nil)
	// Player2 casts Deflecting Palm against Player1's source.
	resolveInstruction(engine, g, &game.StackObject{Controller: game.Player2, SourceID: palmID}, game.PreventDamage{
		Player:                              game.ControllerReference(),
		All:                                 true,
		OneShot:                             true,
		RedirectPreventedToSourceController: true,
	}, nil)

	if dealt := dealPlayerDamage(g, sourceID, 0, game.Player1, game.Player2, 5, false); dealt != 0 {
		t.Fatalf("dealt damage to shielded player = %d, want 0", dealt)
	}
	if g.Players[game.Player2].Life != 20 {
		t.Fatalf("shielded player life = %d, want 20", g.Players[game.Player2].Life)
	}
	if g.Players[game.Player1].Life != 20 {
		t.Fatalf("source controller life = %d, want 20 after redirect prevented", g.Players[game.Player1].Life)
	}
	if len(g.PreventionShields) != 0 {
		t.Fatalf("prevention shields = %+v, want both expired", g.PreventionShields)
	}
}
