package rules

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

// regModifyEffect builds a layer-7c modify continuous effect on one permanent
// with the given duration. The effectID also seeds the timestamp.
func regModifyEffect(effectID id.ID, permanent *game.Permanent, powerDelta, toughnessDelta int, duration game.EffectDuration, createdTurn int) game.ContinuousEffect {
	return game.ContinuousEffect{
		ID:               effectID,
		AffectedObjectID: permanent.ObjectID,
		Timestamp:        game.Timestamp(effectID),
		Layer:            game.LayerPowerToughnessModify,
		Duration:         duration,
		CreatedTurn:      createdTurn,
		PowerDelta:       powerDelta,
		ToughnessDelta:   toughnessDelta,
	}
}

// TestRegUntilEndOfTurnEffectExpiresAtCleanup asserts that an "until end of
// turn" continuous effect applies before cleanup and stops applying after the
// cleanup-step expiry runs.
func TestRegUntilEndOfTurnEffectExpiresAtCleanup(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.ContinuousEffects = append(g.ContinuousEffects,
		regModifyEffect(1, creature, 3, 3, game.DurationUntilEndOfTurn, g.Turn.TurnNumber))

	if got := effectivePower(g, creature); got != 5 {
		t.Fatalf("effective power before cleanup = %d, want 5", got)
	}

	expireCleanupDurations(g)

	if len(g.ContinuousEffects) != 0 {
		t.Fatalf("continuous effects after cleanup = %d, want 0", len(g.ContinuousEffects))
	}
	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("effective power after cleanup = %d, want base 2", got)
	}
}

// TestRegThisTurnEffectExpiresAtCleanup asserts the "this turn" duration also
// expires at cleanup.
func TestRegThisTurnEffectExpiresAtCleanup(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.ContinuousEffects = append(g.ContinuousEffects,
		regModifyEffect(1, creature, 4, 4, game.DurationThisTurn, g.Turn.TurnNumber))

	if got := effectivePower(g, creature); got != 6 {
		t.Fatalf("effective power before cleanup = %d, want 6", got)
	}

	expireCleanupDurations(g)

	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("effective power after cleanup = %d, want base 2", got)
	}
}

// TestRegPermanentDurationSurvivesCleanup asserts that an effect with no
// turn-bound duration persists past the cleanup-step expiry.
func TestRegPermanentDurationSurvivesCleanup(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.ContinuousEffects = append(g.ContinuousEffects,
		regModifyEffect(1, creature, 3, 3, game.DurationPermanent, g.Turn.TurnNumber))

	expireCleanupDurations(g)

	if len(g.ContinuousEffects) != 1 {
		t.Fatalf("permanent-duration effects after cleanup = %d, want 1", len(g.ContinuousEffects))
	}
	if got := effectivePower(g, creature); got != 5 {
		t.Fatalf("effective power after cleanup = %d, want 5 (still applying)", got)
	}
}

// TestRegUntilEndOfTurnEffectExpiresOverTurnBoundary drives the real ending
// phase and asserts the effect is gone after the cleanup step runs.
func TestRegUntilEndOfTurnEffectExpiresOverTurnBoundary(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.ContinuousEffects = append(g.ContinuousEffects,
		regModifyEffect(1, creature, 3, 3, game.DurationUntilEndOfTurn, g.Turn.TurnNumber))

	if got := effectivePower(g, creature); got != 5 {
		t.Fatalf("effective power during turn = %d, want 5", got)
	}

	NewEngine(nil).runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if len(g.ContinuousEffects) != 0 {
		t.Fatalf("continuous effects after ending phase = %d, want 0", len(g.ContinuousEffects))
	}
	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("effective power after ending phase = %d, want base 2", got)
	}
}

// TestRegUntilYourNextTurnExpiresOnlyOnALaterTurn asserts that an "until your
// next turn" effect expires when its controller's later turn begins, but not on
// the turn it was created.
func TestRegUntilYourNextTurnExpiresOnlyOnALaterTurn(t *testing.T) {
	t.Parallel()
	t.Run("created earlier turn expires", func(t *testing.T) {
		t.Parallel()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		g.Turn.ActivePlayer = game.Player1
		g.Turn.TurnNumber = 5
		creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
			ID:               1,
			AffectedObjectID: creature.ObjectID,
			Layer:            game.LayerPowerToughnessModify,
			Duration:         game.DurationUntilYourNextTurn,
			ExpiresFor:       game.Player1,
			CreatedTurn:      1,
			PowerDelta:       3,
			ToughnessDelta:   3,
		})

		expireTurnStartDurations(g)

		if len(g.ContinuousEffects) != 0 {
			t.Fatalf("effects after the controller's later turn began = %d, want 0", len(g.ContinuousEffects))
		}
	})
	t.Run("created this turn persists", func(t *testing.T) {
		t.Parallel()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		g.Turn.ActivePlayer = game.Player1
		g.Turn.TurnNumber = 5
		creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
			ID:               1,
			AffectedObjectID: creature.ObjectID,
			Layer:            game.LayerPowerToughnessModify,
			Duration:         game.DurationUntilYourNextTurn,
			ExpiresFor:       game.Player1,
			CreatedTurn:      5,
			PowerDelta:       3,
			ToughnessDelta:   3,
		})

		expireTurnStartDurations(g)

		if len(g.ContinuousEffects) != 1 {
			t.Fatalf("same-turn effect was expired = %d effects, want 1", len(g.ContinuousEffects))
		}
	})
	t.Run("other player's turn does not expire it", func(t *testing.T) {
		t.Parallel()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		g.Turn.ActivePlayer = game.Player2
		g.Turn.TurnNumber = 5
		creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
			ID:               1,
			AffectedObjectID: creature.ObjectID,
			Layer:            game.LayerPowerToughnessModify,
			Duration:         game.DurationUntilYourNextTurn,
			ExpiresFor:       game.Player1,
			CreatedTurn:      1,
			PowerDelta:       3,
			ToughnessDelta:   3,
		})

		expireTurnStartDurations(g)

		if len(g.ContinuousEffects) != 1 {
			t.Fatalf("effect expired during a different player's turn = %d effects, want 1", len(g.ContinuousEffects))
		}
	})
}

// TestRegUntilEndOfTurnExpiryProperty fuzzes random modify deltas and asserts
// the effect applies before cleanup and never afterward.
func TestRegUntilEndOfTurnExpiryProperty(t *testing.T) {
	t.Parallel()
	rng := rand.New(rand.NewPCG(610, 2))
	for iteration := range 300 {
		basePower := 1 + rng.IntN(5)
		powerDelta := rng.IntN(6)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		creature := addCombatCreaturePermanentWithPower(g, game.Player1, basePower)
		g.ContinuousEffects = append(g.ContinuousEffects,
			regModifyEffect(1, creature, powerDelta, powerDelta, game.DurationUntilEndOfTurn, g.Turn.TurnNumber))

		if got := effectivePower(g, creature); got != basePower+powerDelta {
			t.Fatalf("iteration %d: power before cleanup = %d, want %d", iteration, got, basePower+powerDelta)
		}

		expireCleanupDurations(g)

		if got := effectivePower(g, creature); got != basePower {
			t.Fatalf("iteration %d: power after cleanup = %d, want base %d", iteration, got, basePower)
		}
	}
}
