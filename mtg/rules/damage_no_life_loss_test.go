package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestDamageDoesntCauseLifeLossSkipsLifeLossKeepsDamage models Archon of
// Coronation's "damage doesn't cause you to lose life": a player with an active
// RuleEffectDamageDoesntCauseLifeLoss takes damage (it is still dealt and returned
// as dealt, so combat-damage triggers and the monarch transfer still happen) but
// their life total does not change. A player without the effect loses life
// normally.
func TestDamageDoesntCauseLifeLossSkipsLifeLossKeepsDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind:           game.RuleEffectDamageDoesntCauseLifeLoss,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		Duration:       game.DurationPermanent,
		CreatedTurn:    g.Turn.TurnNumber,
	})

	protectedBefore := g.Players[game.Player1].Life
	dealt := dealPlayerDamage(g, 0, 0, game.Player2, game.Player1, 5, true)
	if dealt != 5 {
		t.Fatalf("damage dealt = %d, want 5 (damage is still dealt)", dealt)
	}
	if g.Players[game.Player1].Life != protectedBefore {
		t.Fatalf("protected player life changed by %d, want 0", protectedBefore-g.Players[game.Player1].Life)
	}

	opponentBefore := g.Players[game.Player2].Life
	if got := dealPlayerDamage(g, 0, 0, game.Player1, game.Player2, 5, true); got != 5 {
		t.Fatalf("unprotected damage dealt = %d, want 5", got)
	}
	if lost := opponentBefore - g.Players[game.Player2].Life; lost != 5 {
		t.Fatalf("unprotected player life lost = %d, want 5", lost)
	}
}
