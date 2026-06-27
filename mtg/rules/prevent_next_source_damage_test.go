package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
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
