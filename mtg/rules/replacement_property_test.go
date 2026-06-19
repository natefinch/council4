package rules

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
)

// regPlayerShield builds an until-end-of-turn prevention shield protecting a
// player from the given amount of damage.
func regPlayerShield(shieldID id.ID, player game.PlayerID, amount int) game.PreventionShield {
	return game.PreventionShield{
		ID:       shieldID,
		Player:   player,
		Amount:   amount,
		Duration: game.DurationUntilEndOfTurn,
	}
}

// regDamageReplacement builds a runtime damage replacement effect that matches
// any damage event and applies the given multiplier and addend.
func regDamageReplacement(effectID id.ID, multiplier, addend int) game.ReplacementEffect {
	return game.ReplacementEffect{
		ID:               effectID,
		MatchEvent:       game.EventDamageDealt,
		ControllerFilter: game.TriggerControllerAny,
		Duration:         game.DurationPermanent,
		DamageMultiplier: multiplier,
		DamageAddend:     addend,
	}
}

// regPlayerDamageEvent builds a damage event aimed at a player.
func regPlayerDamageEvent(player game.PlayerID, amount int) damageEvent {
	return damageEvent{
		controller: game.Player1,
		player:     player,
		amount:     amount,
	}
}

// TestRegPreventionRunsBeforeReplacement asserts the engine applies prevention
// shields first, then damage-increasing replacements, matching CR 616/615
// ordering.
func TestRegPreventionRunsBeforeReplacement(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	sourceID := addColoredSourceCard(g, game.Player1, color.Red)
	addReplacementPermanent(t, g, game.Player1, damageMultiplierReplacementCardDef())
	g.PreventionShields = append(g.PreventionShields, regPlayerShield(1, game.Player2, 2))

	// 5 damage, prevent 2 -> 3, then doubled -> 6. Doubling first would give 8.
	dealt := dealPlayerDamage(g, sourceID, 0, game.Player1, game.Player2, 5, false)
	if dealt != 6 {
		t.Fatalf("dealt = %d, want prevention-then-double 6", dealt)
	}
}

// TestRegPreventionShieldsStackAndChoiceIsRecorded asserts multiple prevention
// shields combine and that an ordering decision is recorded for the damaged
// player when more than one shield applies.
func TestRegPreventionShieldsStackAndChoiceIsRecorded(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.PreventionShields = append(g.PreventionShields,
		regPlayerShield(1, game.Player2, 2),
		regPlayerShield(2, game.Player2, 3),
	)
	event := regPlayerDamageEvent(game.Player2, 4)

	remaining := applyDamagePrevention(g, event)

	// 4 damage against 2+3 prevention -> fully prevented.
	if remaining != 0 {
		t.Fatalf("remaining damage = %d, want 0", remaining)
	}
	if len(g.ReplacementDecisions) != 1 {
		t.Fatalf("replacement decisions = %d, want 1 ordering choice", len(g.ReplacementDecisions))
	}
	if got := g.ReplacementDecisions[0].Player; got != game.Player2 {
		t.Fatalf("ordering decision player = %v, want damaged Player2", got)
	}
}

// TestRegStackedDamageReplacementsRecordOneChoicePerRound asserts that competing
// damage replacements are applied via a recorded fallback decision and produce a
// deterministic add-then-double result.
func TestRegStackedDamageReplacementsRecordOneChoicePerRound(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.ReplacementEffects = append(g.ReplacementEffects,
		regDamageReplacement(1, 1, 2), // +2 addend, listed first
		regDamageReplacement(2, 2, 0), // doubling
	)
	event := regPlayerDamageEvent(game.Player2, 3)

	got := replacementDamageAmount(g, event)

	// Round 1 picks the first match (+2) -> 5; round 2 doubles -> 10.
	if got != 10 {
		t.Fatalf("replaced damage = %d, want add-then-double 10", got)
	}
	if len(g.ReplacementDecisions) != 1 {
		t.Fatalf("replacement decisions = %d, want exactly one for the contested round", len(g.ReplacementDecisions))
	}
	if g.ReplacementDecisions[0].Player != game.Player2 {
		t.Fatalf("decision player = %v, want damaged Player2", g.ReplacementDecisions[0].Player)
	}
}

// TestRegPreventionShieldStackingProperty fuzzes random shields and damage and
// asserts the remaining damage equals the floored difference, never negative.
func TestRegPreventionShieldStackingProperty(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewPCG(610, 3))
	for iteration := range 400 {
		shieldCount := rng.IntN(4)
		totalPrevention := 0
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		for range shieldCount {
			amount := 1 + rng.IntN(5)
			totalPrevention += amount
			g.PreventionShields = append(g.PreventionShields, regPlayerShield(g.IDGen.Next(), game.Player2, amount))
		}
		damage := 1 + rng.IntN(12)
		event := regPlayerDamageEvent(game.Player2, damage)

		remaining := applyDamagePrevention(g, event)

		want := max(damage-totalPrevention, 0)
		if remaining != want {
			t.Fatalf("iteration %d: remaining = %d, want %d (damage %d shields %d)",
				iteration, remaining, want, damage, totalPrevention)
		}
		if remaining < 0 {
			t.Fatalf("iteration %d: remaining damage went negative: %d", iteration, remaining)
		}
	}
}

// TestRegPreventionThenDoublingProperty fuzzes random prevention and damage with
// a fixed doubling replacement and asserts prevention always precedes doubling.
func TestRegPreventionThenDoublingProperty(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewPCG(610, 4))
	for iteration := range 300 {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		sourceID := addColoredSourceCard(g, game.Player1, color.Red)
		addReplacementPermanent(t, g, game.Player1, damageMultiplierReplacementCardDef())

		damage := 1 + rng.IntN(10)
		prevention := rng.IntN(damage + 3)
		if prevention > 0 {
			g.PreventionShields = append(g.PreventionShields, regPlayerShield(g.IDGen.Next(), game.Player2, prevention))
		}

		startLife := g.Players[game.Player2].Life
		dealt := dealPlayerDamage(g, sourceID, 0, game.Player1, game.Player2, damage, false)

		afterPrevention := max(damage-prevention, 0)
		want := afterPrevention * 2
		if dealt != want {
			t.Fatalf("iteration %d: dealt = %d, want prevent-then-double %d (damage %d prevention %d)",
				iteration, dealt, want, damage, prevention)
		}
		if got := startLife - g.Players[game.Player2].Life; got != want {
			t.Fatalf("iteration %d: life lost = %d, want %d", iteration, got, want)
		}
	}
}
