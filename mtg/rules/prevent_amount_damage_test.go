package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
)

// TestPreventAmountAnyTargetShieldsChosenPermanent proves the amount-based
// any-target prevention shield ("Prevent the next N damage that would be dealt
// to any target this turn.") resolves to the permanent its target slot named,
// preventing N damage dealt to that creature.
func TestPreventAmountAnyTargetShieldsChosenPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	obj := &game.StackObject{
		Controller: game.Player2,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}

	resolveInstruction(engine, g, obj, game.PreventDamage{
		Amount:    game.Fixed(2),
		AnyTarget: game.AnyTargetDamageRecipient(0),
	}, nil)

	dealt := dealPermanentDamage(g, sourceID, 0, game.Player1, target, 5, false)
	if dealt != 3 {
		t.Fatalf("dealt damage = %d, want 3 after any-target prevention shield", dealt)
	}
	if target.MarkedDamage != 3 {
		t.Fatalf("marked damage = %d, want 3", target.MarkedDamage)
	}
	if len(g.PreventionShields) != 0 {
		t.Fatalf("prevention shields = %+v, want consumed", g.PreventionShields)
	}
	assertEvent(t, g.Events, game.EventDamagePrevented, func(event game.Event) bool {
		return event.PermanentID == target.ObjectID && event.Amount == 2
	})
}

// TestPreventAmountAnyTargetShieldsChosenPlayer proves the same shield resolves
// to the player its target slot named, preventing N damage dealt to that player.
func TestPreventAmountAnyTargetShieldsChosenPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Players[game.Player2].Life = 20
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)
	obj := &game.StackObject{
		Controller: game.Player2,
		Targets:    []game.Target{game.PlayerTarget(game.Player2)},
	}

	resolveInstruction(engine, g, obj, game.PreventDamage{
		Amount:    game.Fixed(2),
		AnyTarget: game.AnyTargetDamageRecipient(0),
	}, nil)

	dealt := dealPlayerDamage(g, sourceID, 0, game.Player1, game.Player2, 5, false)
	if dealt != 3 {
		t.Fatalf("dealt damage = %d, want 3 after any-target prevention shield", dealt)
	}
	if g.Players[game.Player2].Life != 17 {
		t.Fatalf("player life = %d, want 17", g.Players[game.Player2].Life)
	}
	if len(g.PreventionShields) != 0 {
		t.Fatalf("prevention shields = %+v, want consumed", g.PreventionShields)
	}
	assertEvent(t, g.Events, game.EventDamagePrevented, func(event game.Event) bool {
		return event.Player == game.Player2 && event.Amount == 2
	})
}
