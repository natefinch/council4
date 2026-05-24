package game

import (
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
)

// CounterPlacement describes counters a permanent enters with.
type CounterPlacement struct {
	Kind   counter.Kind
	Amount int
}

// PreventionShield prevents an amount of future damage to a player or
// permanent.
type PreventionShield struct {
	ID          id.ID
	Controller  PlayerID
	Player      PlayerID
	PermanentID id.ID
	Amount      int
	Duration    EffectDuration
	CreatedTurn int
}

// ReplacementDecision records deterministic ordering for competing replacement
// or prevention effects.
type ReplacementDecision struct {
	Player       PlayerID
	Options      []string
	Selected     []int
	UsedFallback bool
}
