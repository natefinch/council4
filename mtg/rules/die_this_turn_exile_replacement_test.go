package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestDieThisTurnExileRedirectsLethalDamageDeath proves the end-to-end behavior
// of the "If that creature would die this turn, exile it instead." rider that
// cardgen lowers onto single-target damage spells such as Lava Coil: the
// CreateReplacement bound to the spell's target permanent redirects that
// permanent's lethal-damage death from the graveyard to exile, while an
// untargeted creature dying the same way still reaches the graveyard.
func TestDieThisTurnExileRedirectsLethalDamageDeath(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	bystander := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	// Resolve the replacement exactly as cardgen emits it: object bound to the
	// spell's first target permanent, redirecting battlefield -> graveyard to
	// exile for the turn.
	resolveInstruction(engine, g, &game.StackObject{
		Kind:       game.StackSpell,
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}, game.CreateReplacement{
		Object:   game.TargetPermanentReference(0),
		Duration: game.DurationThisTurn,
		Replacement: &game.ReplacementEffect{
			Description:   "die-this-turn exile",
			MatchEvent:    game.EventZoneChanged,
			MatchFromZone: true,
			FromZone:      zone.Battlefield,
			MatchToZone:   true,
			ToZone:        zone.Graveyard,
			ReplaceToZone: zone.Exile,
		},
	}, nil)

	if len(g.ReplacementEffects) != 1 {
		t.Fatalf("replacement effects = %d, want 1", len(g.ReplacementEffects))
	}
	if got := g.ReplacementEffects[0].AffectedObjectID; got != target.ObjectID {
		t.Fatalf("AffectedObjectID = %v, want %v", got, target.ObjectID)
	}

	// Both creatures take lethal damage; only the targeted one is redirected.
	target.MarkedDamage = 2
	bystander.MarkedDamage = 2

	changed, _ := engine.checkPermanentStateBasedActions(g, newPassBatchID(g))
	if !changed {
		t.Fatal("checkPermanentStateBasedActions() = false, want lethal-damage deaths")
	}

	if !g.Players[game.Player2].Exile.Contains(target.CardInstanceID) {
		t.Fatal("targeted creature should be exiled instead of going to the graveyard")
	}
	if g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("targeted creature should not be in the graveyard")
	}
	if !g.Players[game.Player2].Graveyard.Contains(bystander.CardInstanceID) {
		t.Fatal("untargeted creature should reach the graveyard, not be redirected")
	}
	if g.Players[game.Player2].Exile.Contains(bystander.CardInstanceID) {
		t.Fatal("untargeted creature should not be exiled")
	}
}

// TestDieThisTurnExileExpiresAtEndOfTurn proves the redirect is scoped to the
// turn it is created: after the this-turn replacement expires at cleanup, a
// later lethal-damage death of the same permanent reaches the graveyard
// normally.
func TestDieThisTurnExileExpiresAtEndOfTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	resolveInstruction(engine, g, &game.StackObject{
		Kind:       game.StackSpell,
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(target.ObjectID)},
	}, game.CreateReplacement{
		Object:   game.TargetPermanentReference(0),
		Duration: game.DurationThisTurn,
		Replacement: &game.ReplacementEffect{
			Description:   "die-this-turn exile",
			MatchEvent:    game.EventZoneChanged,
			MatchFromZone: true,
			FromZone:      zone.Battlefield,
			MatchToZone:   true,
			ToZone:        zone.Graveyard,
			ReplaceToZone: zone.Exile,
		},
	}, nil)
	if len(g.ReplacementEffects) != 1 {
		t.Fatalf("replacement effects = %d, want 1", len(g.ReplacementEffects))
	}

	expireReplacementEffects(g)
	if len(g.ReplacementEffects) != 0 {
		t.Fatalf("replacement effects after cleanup = %d, want 0", len(g.ReplacementEffects))
	}

	target.MarkedDamage = 2
	if _, _ = engine.checkPermanentStateBasedActions(g, newPassBatchID(g)); !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("after the rider expires, lethal damage should send the creature to the graveyard")
	}
	if g.Players[game.Player2].Exile.Contains(target.CardInstanceID) {
		t.Fatal("creature should not be exiled after the rider expires")
	}
}
