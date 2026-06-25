package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLeaveBattlefieldExileReplacementRedirectsAffectedObject verifies that the
// leaves-the-battlefield exile replacement created by CreateReplacement with a
// resolved Object ("If it would leave the battlefield, exile it instead of
// putting it anywhere else." — Whip of Erebos) is bound to that single object:
// the affected permanent is exiled instead of going to the graveyard, while a
// different creature leaving the battlefield is unaffected.
func TestLeaveBattlefieldExileReplacementRedirectsAffectedObject(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	affected := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	other := addCombatCreaturePermanentWithPower(g, game.Player1, 2)

	resolveInstruction(engine, g, &game.StackObject{
		Kind:       game.StackActivatedAbility,
		Controller: game.Player1,
		SourceID:   affected.ObjectID,
	}, game.CreateReplacement{
		Object: game.SourcePermanentReference(),
		Replacement: &game.ReplacementEffect{
			Description:   "exile instead",
			MatchEvent:    game.EventZoneChanged,
			MatchFromZone: true,
			FromZone:      zone.Battlefield,
			ReplaceToZone: zone.Exile,
		},
	}, nil)

	if len(g.ReplacementEffects) != 1 {
		t.Fatalf("replacement effects = %d, want 1", len(g.ReplacementEffects))
	}
	if got := g.ReplacementEffects[0].AffectedObjectID; got != affected.ObjectID {
		t.Fatalf("AffectedObjectID = %v, want %v", got, affected.ObjectID)
	}

	if !movePermanentToZone(g, other, zone.Graveyard) {
		t.Fatal("movePermanentToZone(other) = false, want true")
	}
	if !g.Players[game.Player1].Graveyard.Contains(other.CardInstanceID) {
		t.Fatal("unaffected creature should reach the graveyard, not be redirected")
	}

	if !movePermanentToZone(g, affected, zone.Graveyard) {
		t.Fatal("movePermanentToZone(affected) = false, want true")
	}
	if !g.Players[game.Player1].Exile.Contains(affected.CardInstanceID) {
		t.Fatal("affected creature should be exiled instead of going to the graveyard")
	}
	if g.Players[game.Player1].Graveyard.Contains(affected.CardInstanceID) {
		t.Fatal("affected creature should not be in the graveyard")
	}
}

// TestLeaveBattlefieldExileReplacementFailsClosedWhenObjectUnresolved verifies
// that CreateReplacement with an Object that cannot resolve registers no
// replacement and fails closed.
func TestLeaveBattlefieldExileReplacementFailsClosedWhenObjectUnresolved(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	resolveInstruction(engine, g, &game.StackObject{
		Kind:       game.StackActivatedAbility,
		Controller: game.Player1,
	}, game.CreateReplacement{
		Object: game.SourcePermanentReference(),
		Replacement: &game.ReplacementEffect{
			MatchEvent:    game.EventZoneChanged,
			MatchFromZone: true,
			FromZone:      zone.Battlefield,
			ReplaceToZone: zone.Exile,
		},
	}, nil)

	if len(g.ReplacementEffects) != 0 {
		t.Fatalf("replacement effects = %d, want 0 (object unresolved)", len(g.ReplacementEffects))
	}
}
