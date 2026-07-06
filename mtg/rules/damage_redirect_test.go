package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestRedirectDamageToSourcePreventsMonarchTheft proves that redirected combat
// damage does not reach the defending player, so an attacker cannot steal the
// monarchy (Protector of the Crown's purpose) and the player takes no commander
// damage — the combat damage is dealt to the redirect creature instead.
func TestRedirectDamageToSourcePreventsMonarchTheft(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMonarch(g, game.Player1)
	protector := addCreaturePermanent(g, game.Player1)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind:           game.RuleEffectRedirectDamageToSource,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		SourceObjectID: protector.ObjectID,
		Duration:       game.DurationPermanent,
		CreatedTurn:    g.Turn.TurnNumber,
	})
	attacker := addCreaturePermanent(g, game.Player2)

	lifeBefore := g.Players[game.Player1].Life
	markPlayerCombatDamage(g, attacker, game.Player1, 3, &TurnLog{})

	if !g.Players[game.Player1].IsMonarch {
		t.Fatal("monarchy was stolen by redirected combat damage, want retained")
	}
	if g.Players[game.Player1].Life != lifeBefore {
		t.Fatalf("player life changed by %d, want 0 (damage redirected)", lifeBefore-g.Players[game.Player1].Life)
	}
	if protector.MarkedDamage != 3 {
		t.Fatalf("redirect creature marked damage = %d, want 3", protector.MarkedDamage)
	}
}

// damage that would be dealt to you is dealt to this creature instead": damage
// aimed at a player with an active RuleEffectRedirectDamageToSource is dealt to
// the effect's source permanent instead, so the player's life is unchanged and
// the creature is marked with the damage. A player without the effect loses life
// normally.
func TestRedirectDamageToSourceDealsToCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	protector := addCreaturePermanent(g, game.Player1)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind:           game.RuleEffectRedirectDamageToSource,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		SourceObjectID: protector.ObjectID,
		Duration:       game.DurationPermanent,
		CreatedTurn:    g.Turn.TurnNumber,
	})

	lifeBefore := g.Players[game.Player1].Life
	dealt := dealPlayerDamage(g, 0, 0, game.Player2, game.Player1, 5, true)
	if dealt != 5 {
		t.Fatalf("damage dealt = %d, want 5 (redirected damage is still dealt)", dealt)
	}
	if g.Players[game.Player1].Life != lifeBefore {
		t.Fatalf("protected player life changed by %d, want 0 (damage redirected)", lifeBefore-g.Players[game.Player1].Life)
	}
	if protector.MarkedDamage != 5 {
		t.Fatalf("redirect creature marked damage = %d, want 5", protector.MarkedDamage)
	}

	opponentBefore := g.Players[game.Player2].Life
	if got := dealPlayerDamage(g, 0, 0, game.Player1, game.Player2, 3, true); got != 3 {
		t.Fatalf("unredirected damage dealt = %d, want 3", got)
	}
	if lost := opponentBefore - g.Players[game.Player2].Life; lost != 3 {
		t.Fatalf("unredirected player life lost = %d, want 3", lost)
	}
}
