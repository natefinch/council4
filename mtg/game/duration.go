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
	// DurationUntilEndOfCombat expires when the combat phase ends ("this combat",
	// Canal Courier's "this creature can't be blocked this combat"). It is removed
	// as the combat phase is torn down, before the following phases. Added last so
	// existing durations keep their wire values.
	DurationUntilEndOfCombat
	// DurationForAsLongAsPlayerIsMonarch expires when the player bound in
	// ExpiresFor is no longer the monarch ("gain control of target creature that
	// player controls for as long as they're the monarch.", Garland, Royal
	// Kidnapper). The bound player is the one whose becoming the monarch created
	// the effect; when a different player takes the crown, or no player is the
	// monarch, the effect ends. Added last so existing durations keep their wire
	// values.
	DurationForAsLongAsPlayerIsMonarch
)

// DelayedTriggerTiming describes when a delayed triggered ability should fire.
type DelayedTriggerTiming int

// Delayed trigger timing values identify supported delayed trigger windows.
const (
	DelayedAtBeginningOfNextEndStep DelayedTriggerTiming = iota + 1
	DelayedAtBeginningOfNextUpkeep
	DelayedAtBeginningOfNextMainPhase
	DelayedAtEndOfCombat
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
	// DelayedWindowUntilFires bounds an event-based delayed trigger to its own
	// firing: it persists across turns until a matching event occurs, then a
	// OneShot trigger removes itself. It is never removed at turn cleanup, backing
	// return conditions with no turn bound ("until an opponent becomes the
	// monarch", Palace Jailer). It is only valid with OneShot set.
	DelayedWindowUntilFires
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
	// DamageSourceObject, when present, binds an EventPattern combat-damage
	// delayed trigger to the specific permanent this object reference resolves
	// to at schedule time (the creature an earlier clause in the same resolution
	// targeted and published as a linked object). The scheduled trigger fires
	// only on combat damage dealt by that captured permanent ("... target
	// creature ... Whenever that creature deals combat damage to a player this
	// turn, ..."). It is only valid with EventPattern whose DamageSourceCaptured
	// flag is set.
	DamageSourceObject opt.V[ObjectReference]
	// CapturedObject, when present, freezes the permanent this object reference
	// resolves to against the creating ability's triggering event at schedule
	// time, storing its object ID on the scheduled trigger so the trigger's
	// content can act on it once the original event is gone. It backs delayed
	// "at end of combat" disposal of the creature involved in combat ("destroy
	// that creature at end of combat"), where CapturedObject is the event or
	// event-related permanent and the content references
	// ObjectReferenceCapturedObject. It is only valid with a fixed-phase Timing.
	CapturedObject opt.V[ObjectReference]
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
	// BoundDamageSourceObjectID restricts an EventPattern combat-damage delayed
	// trigger whose pattern sets DamageSourceCaptured to events whose damage
	// source is this captured permanent. It is resolved from the creating
	// ability's DamageSourceObject reference when the trigger is scheduled. Zero
	// means the captured permanent was already gone, so the trigger never fires.
	BoundDamageSourceObjectID id.ID
	// CapturedObjectID is the concrete permanent frozen from the creating
	// ability's CapturedObject reference at schedule time, carried into the
	// fired trigger's content so it can act on the combat creature after the
	// original event is gone ("destroy that creature at end of combat"). Zero
	// means no object was captured, so content that references
	// ObjectReferenceCapturedObject finds nothing and does nothing.
	CapturedObjectID id.ID
}
