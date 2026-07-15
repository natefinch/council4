package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// gigglingSkitterspikeDamage is the runtime shape the Oracle backend lowers each
// of Giggling Skitterspike's three combat/target triggers to: deal damage equal
// to the triggering permanent's power to each opponent, sourced from that
// permanent so its keywords apply and its last-known power is used once it has
// left the battlefield.
func gigglingSkitterspikeDamage() game.Damage {
	return game.Damage{
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:       game.DynamicAmountObjectPower,
			Multiplier: 1,
			Object:     game.EventPermanentReference(),
		}),
		Recipient:    game.PlayerGroupDamageRecipient(game.OpponentsReference()),
		DamageSource: opt.Val(game.EventPermanentReference()),
	}
}

// TestGigglingSkitterspikeDamageHitsEachOpponent proves the source-power
// group-damage body deals full source-power damage separately to every opponent
// in a multiplayer game, leaving the controller untouched.
func TestGigglingSkitterspikeDamageHitsEachOpponent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 4)

	start := [game.NumPlayers]int{}
	for p := range game.NumPlayers {
		start[p] = g.Players[p].Life
	}

	obj := &game.StackObject{
		ID:              g.IDGen.Next(),
		Controller:      game.Player1,
		SourceID:        source.ObjectID,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:        game.EventAttackerDeclared,
			Controller:  game.Player1,
			PermanentID: source.ObjectID,
		},
	}
	resolveInstruction(engine, g, obj, gigglingSkitterspikeDamage(), &TurnLog{})

	if got := g.Players[game.Player1].Life; got != start[game.Player1] {
		t.Fatalf("controller life = %d, want %d (controller is not an opponent)", got, start[game.Player1])
	}
	for _, opponent := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if got := start[opponent] - g.Players[opponent].Life; got != 4 {
			t.Fatalf("opponent %d life lost = %d, want 4 (full source power dealt separately)", opponent, got)
		}
	}
}

// TestGigglingSkitterspikeDamageUsesLastKnownWhenSourceLeaves proves the trigger
// still deals damage after its source has left the battlefield before the
// ability resolves: the amount reads the source's last-known power (CR 608.2h)
// and the damage source's last-known lifelink still gains its controller life
// (once per opponent dealt damage).
func TestGigglingSkitterspikeDamageUsesLastKnownWhenSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 6, game.Lifelink)

	start := [game.NumPlayers]int{}
	for p := range game.NumPlayers {
		start[p] = g.Players[p].Life
	}

	obj := &game.StackObject{
		ID:              g.IDGen.Next(),
		Controller:      game.Player1,
		SourceID:        source.ObjectID,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:        game.EventAttackerDeclared,
			Controller:  game.Player1,
			PermanentID: source.ObjectID,
		},
	}

	// The source leaves the battlefield before the trigger resolves; its
	// last-known information is what the resolving ability must read.
	snapshot := snapshotPermanent(g, source, zone.Battlefield)
	removePermanentFromBattlefield(g, source.ObjectID)
	rememberLastKnown(g, &snapshot)

	resolveInstruction(engine, g, obj, gigglingSkitterspikeDamage(), &TurnLog{})

	for _, opponent := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if got := start[opponent] - g.Players[opponent].Life; got != 6 {
			t.Fatalf("opponent %d life lost = %d, want 6 (last-known source power)", opponent, got)
		}
	}
	// Lifelink from the departed source: 6 damage to each of three opponents.
	if got := g.Players[game.Player1].Life - start[game.Player1]; got != 18 {
		t.Fatalf("controller life gained = %d, want 18 (last-known lifelink, 6 per opponent)", got)
	}
}
