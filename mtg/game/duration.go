package game

import (
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/opt"
)

// EffectDuration describes when a runtime continuous effect expires.
type EffectDuration int

// Effect duration values identify when runtime effects expire.
const (
	DurationPermanent EffectDuration = iota
	DurationUntilEndOfTurn
	DurationThisTurn
	DurationUntilYourNextTurn
	DurationNextTime
	DurationUntilEndOfYourNextTurn
	// DurationForAsLongAsSourceOnBattlefield expires when the source permanent
	// is no longer on the battlefield. Use object identity (SourceObjectID),
	// never card name, to identify the source.
	DurationForAsLongAsSourceOnBattlefield
	// DurationForAsLongAsYouControlSource expires when the effect controller no
	// longer controls the source permanent, or when the source permanent leaves
	// the battlefield. Use object identity (SourceObjectID), never card name.
	DurationForAsLongAsYouControlSource
	// DurationForAsLongAsControlledCreatureEnchanted expires when the affected
	// permanent is no longer enchanted (no Aura attached) or has left the
	// battlefield. It models attachment-dependent control durations such as
	// "for as long as that creature is enchanted". Use object identity
	// (AffectedObjectID), never card name.
	DurationForAsLongAsControlledCreatureEnchanted
	// DurationUntilYourNextEndStep expires at the controller's next end step
	// ("until your next end step", Inti, Seneschal of the Sun). It is removed at
	// the cleanup following that end step.
	DurationUntilYourNextEndStep
)

// DelayedTriggerTiming describes when a delayed triggered ability should fire.
type DelayedTriggerTiming int

// Delayed trigger timing values identify supported delayed trigger windows.
const (
	DelayedAtBeginningOfNextEndStep DelayedTriggerTiming = iota + 1
	DelayedAtBeginningOfNextUpkeep
	DelayedAtBeginningOfNextMainPhase
)

// DelayedTriggerWindow bounds the lifetime of an event-based delayed trigger.
type DelayedTriggerWindow int

// Delayed trigger window values identify supported event-trigger windows.
const (
	// DelayedWindowNone is the zero value for fixed-phase delayed triggers.
	DelayedWindowNone DelayedTriggerWindow = iota
	// DelayedWindowThisTurn bounds an event-based delayed trigger to the turn it
	// was created on ("... this turn"). It is removed during that turn's cleanup
	// step.
	DelayedWindowThisTurn
)

// DelayedTriggerDef is the card-definition-side data for creating a delayed
// triggered ability.
type DelayedTriggerDef struct {
	Timing   DelayedTriggerTiming
	Optional bool
	Content  AbilityContent
	// EventPattern, when present, makes this an event-based delayed trigger that
	// fires when a matching game event occurs within its window, reusing the
	// ordinary triggered-ability event matcher. Timing must be zero when
	// EventPattern is present, and Window must be non-zero.
	EventPattern opt.V[TriggerPattern]
	// OneShot, valid only with EventPattern, removes the delayed trigger after it
	// fires once ("the next time you cast ..."). When false the trigger fires on
	// every matching event until its window ends ("whenever you cast ... this
	// turn").
	OneShot bool
	// Window bounds an EventPattern delayed trigger's lifetime. It must be
	// non-zero when EventPattern is present and zero otherwise.
	Window DelayedTriggerWindow
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
	Ability        TriggeredAbility
	// EventPattern, OneShot, and Window mirror the same fields on
	// DelayedTriggerDef for an event-based delayed trigger. EventPattern is
	// absent for fixed-phase delayed triggers, which fire on a step boundary via
	// Timing instead.
	EventPattern opt.V[TriggerPattern]
	OneShot      bool
	Window       DelayedTriggerWindow
	// CapturedTargetControllerLKI preserves target-derived player references
	// captured from the spell or ability that created this delayed trigger.
	CapturedTargetControllerLKI map[int]PlayerID
	// CapturedTargetManaValueLKI preserves target spell mana values captured
	// from the spell or ability that created this delayed trigger.
	CapturedTargetManaValueLKI map[int]int
}
