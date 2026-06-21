package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
)

func TestPreventAllCombatDamageToCreatureShield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	obj := &game.StackObject{
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}

	resolveInstruction(engine, g, obj, game.PreventDamage{
		Object:     game.TargetPermanentReference(0),
		All:        true,
		CombatOnly: true,
	}, nil)

	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, target, 4, true); dealt != 0 {
		t.Fatalf("combat damage to shielded creature = %d, want 0", dealt)
	}
	// The shield has no fixed capacity and persists for further combat damage.
	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, target, 3, true); dealt != 0 {
		t.Fatalf("second combat damage to shielded creature = %d, want 0", dealt)
	}
	// Noncombat damage is unaffected.
	if dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, target, 2, false); dealt != 2 {
		t.Fatalf("noncombat damage to shielded creature = %d, want 2", dealt)
	}
	if target.MarkedDamage != 2 {
		t.Fatalf("marked damage = %d, want 2", target.MarkedDamage)
	}
}

func TestPreventAllCombatDamageByCreatureShield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(source.ObjectID)},
	}

	resolveInstruction(engine, g, obj, game.PreventDamage{
		Object:     game.TargetPermanentReference(0),
		All:        true,
		CombatOnly: true,
		BySource:   true,
	}, nil)

	startLife := g.Players[game.Player2].Life
	if dealt := dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, game.Player1, game.Player2, 4, true); dealt != 0 {
		t.Fatalf("combat damage by shielded creature = %d, want 0", dealt)
	}
	if g.Players[game.Player2].Life != startLife {
		t.Fatalf("player life = %d, want unchanged %d", g.Players[game.Player2].Life, startLife)
	}
	// Damage dealt by a different source is unaffected.
	otherID := addColoredSourceCard(g, game.Player1, color.Red)
	if dealt := dealPlayerDamage(g, otherID, 0, game.Player1, game.Player2, 2, true); dealt != 2 {
		t.Fatalf("combat damage by other source = %d, want 2", dealt)
	}
	// Noncombat damage by the shielded creature is unaffected.
	if dealt := dealPlayerDamage(g, source.CardInstanceID, source.ObjectID, game.Player1, game.Player2, 3, false); dealt != 3 {
		t.Fatalf("noncombat damage by shielded creature = %d, want 3", dealt)
	}
}

// TestPreventAllCombatDamageGlobalShield covers the global combat-damage
// prevention shield (Spike Weaver). Once resolved, no combat damage is dealt
// this turn to any permanent or player, regardless of source, while noncombat
// damage is unaffected.
func TestPreventAllCombatDamageGlobalShield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 4)
	obj := &game.StackObject{Controller: game.Player1}

	resolveInstruction(engine, g, obj, game.PreventDamage{
		All:        true,
		CombatOnly: true,
		Global:     true,
	}, nil)

	// Combat damage between two unrelated creatures is prevented.
	if dealt := dealPermanentDamage(g, attacker.CardInstanceID, attacker.ObjectID, game.Player1, blocker, 5, true); dealt != 0 {
		t.Fatalf("combat damage to creature = %d, want 0", dealt)
	}
	if dealt := dealPermanentDamage(g, blocker.CardInstanceID, blocker.ObjectID, game.Player2, attacker, 4, true); dealt != 0 {
		t.Fatalf("combat damage to attacker = %d, want 0", dealt)
	}
	// Combat damage to a player is also prevented.
	startLife := g.Players[game.Player2].Life
	if dealt := dealPlayerDamage(g, attacker.CardInstanceID, attacker.ObjectID, game.Player1, game.Player2, 5, true); dealt != 0 {
		t.Fatalf("combat damage to player = %d, want 0", dealt)
	}
	if g.Players[game.Player2].Life != startLife {
		t.Fatalf("player life = %d, want unchanged %d", g.Players[game.Player2].Life, startLife)
	}
	// Noncombat damage is unaffected.
	if dealt := dealPermanentDamage(g, attacker.CardInstanceID, attacker.ObjectID, game.Player1, blocker, 2, false); dealt != 2 {
		t.Fatalf("noncombat damage to creature = %d, want 2", dealt)
	}
}
