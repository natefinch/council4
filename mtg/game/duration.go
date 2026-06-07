package game

import "github.com/natefinch/council4/mtg/game/id"

// EffectDuration describes when a runtime continuous effect expires.
type EffectDuration int

// Effect duration values identify when runtime effects expire.
const (
	DurationPermanent EffectDuration = iota
	DurationUntilEndOfTurn
	DurationThisTurn
	DurationUntilYourNextTurn
	DurationNextTime
)

// DelayedTriggerTiming describes when a delayed triggered ability should fire.
type DelayedTriggerTiming int

// Delayed trigger timing values identify supported delayed trigger windows.
const (
	DelayedAtBeginningOfNextEndStep DelayedTriggerTiming = iota + 1
)

// DelayedTriggerDef is the card-definition-side data for creating a delayed
// triggered ability.
type DelayedTriggerDef struct {
	Timing   DelayedTriggerTiming
	Optional bool
	Content  AbilityContent
}

// DelayedTrigger is a runtime delayed triggered ability waiting for its timing
// condition.
type DelayedTrigger struct {
	ID             id.ID
	SourceID       id.ID
	SourceObjectID id.ID
	SourceTokenDef *CardDef
	Controller     PlayerID
	CreatedTurn    int
	Timing         DelayedTriggerTiming
	Ability        TriggeredAbilityBody
}
