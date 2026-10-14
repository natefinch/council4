package game

import (
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ZoneTransitionPolicy is an immutable exception to ordinary zone-change rules.
// It never preserves object identity or authorizes arbitrary state mutation.
type ZoneTransitionPolicy int

const (
	// ZoneTransitionOrdinary uses CR 400.7 and 122.2.
	ZoneTransitionOrdinary ZoneTransitionPolicy = iota
	// ZoneTransitionRetainCountersExceptHandLibrary retains all counters on
	// actual arrivals elsewhere. With no explicit ZoneOfFunction, this ability
	// functions everywhere (CR 113.6c), including off the battlefield.
	ZoneTransitionRetainCountersExceptHandLibrary
)

// ZoneTransitionPolicyProvider exposes typed declarations on effective abilities.
type ZoneTransitionPolicyProvider interface {
	ZoneTransitionDeclaration() (ZoneTransitionPolicy, zone.Type, opt.V[Condition])
}

// ZoneTransitionDeclaration returns the read-only transition policy and scope.
func (a *StaticAbility) ZoneTransitionDeclaration() (ZoneTransitionPolicy, zone.Type, opt.V[Condition]) {
	return a.ZoneTransitionPolicy, a.ZoneOfFunction, a.Condition
}

// StaticContinuousEffectProvider exposes immutable continuous declarations.
type StaticContinuousEffectProvider interface {
	StaticContinuousDeclarations() (zone.Type, opt.V[Condition], []ContinuousEffect)
}

// StaticContinuousDeclarations returns the read-only continuous declarations.
func (a *StaticAbility) StaticContinuousDeclarations() (zone.Type, opt.V[Condition], []ContinuousEffect) {
	return a.ZoneOfFunction, a.Condition, a.ContinuousEffects
}
