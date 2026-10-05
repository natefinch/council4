package game

import "github.com/natefinch/council4/mtg/game/id"

// ActivatedAbilityResolutionUse identifies a captured ability occurrence on a
// source incarnation. Ability is immutable shared rules data, not an effective
// ability index that can shift when the source transforms or merges.
type ActivatedAbilityResolutionUse struct {
	SourceID          id.ID
	SourceZoneVersion uint64
	Ability           *ActivatedAbility
	// OriginID identifies the defining card/component or granting card.
	OriginID id.ID
	// Granted identities include the granting source/effect, not the grant's
	// mutable position in the effective ability list.
	Granted       bool
	GrantID       id.ID
	GrantSourceID id.ID
	// Occurrence distinguishes repeated bodies within one immutable grant/copy.
	Occurrence int
}
