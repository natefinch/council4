package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerDelayedCapturedCombatDisposal lowers the delayed "at end of combat"
// disposal of the creature involved in combat — the basilisk family "Whenever
// this creature blocks or becomes blocked by a creature, destroy that creature
// at end of combat" (Tangle Asp, Serpentine Basilisk, Deathgazer). The clause
// names the blocked, blocking, or combat-damaged creature with a demonstrative
// ("that creature") that the compiler binds to the triggering event's permanent
// or its related permanent. That creature must be remembered across the delay,
// because the combat event is gone when the trigger fires at end of combat, so
// the effect is lowered as a CreateDelayedTrigger whose CapturedObject freezes
// the event permanent at schedule time and whose content acts on the frozen
// permanent through ObjectReferenceCapturedObject.
//
// It handles only the destroy verb, whose non-target immediate lowering the
// executable backend does not otherwise support (destroy requires a target
// spell), and only a self trigger whose single effect is an exact, unconditional
// destruction of that one demonstrative creature. An optional ("you may")
// trigger body keeps its optionality on the enclosing triggered ability (hoisted
// by the compiler for triggered abilities), so it is represented faithfully.
// Everything else falls through to the generic delayed-effect path, which fails
// closed.
func lowerDelayedCapturedCombatDisposal(
	ctx contentCtx,
	timing game.DelayedTriggerTiming,
) (game.AbilityContent, bool) {
	if !ctx.selfTrigger || timing == 0 {
		return game.AbilityContent{}, false
	}
	if len(ctx.content.Effects) != 1 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectDestroy ||
		effect.Negated ||
		effect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	if len(ctx.content.References) != 1 {
		return game.AbilityContent{}, false
	}
	// "that creature" names the combat creature, not the source: a demonstrative
	// bound to the event (or event-related) permanent. A pronoun that denotes the
	// source itself belongs to the self-disposal path and is declined here.
	if referencesDenoteSelf(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	eventRef, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowEvent: true})
	if !ok {
		return game.AbilityContent{}, false
	}
	if eventRef.Kind() != game.ObjectReferenceEventPermanent &&
		eventRef.Kind() != game.ObjectReferenceEventRelatedPermanent {
		return game.AbilityContent{}, false
	}
	consumed := ctx
	consumed.content.References = nil
	if consumed.content.Unconsumed() {
		return game.AbilityContent{}, false
	}
	content := game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Destroy{
			Object:              game.CapturedObjectReference(),
			PreventRegeneration: effect.PreventRegeneration,
		},
	}}}.Ability()
	return game.Mode{Sequence: []game.Instruction{{Primitive: game.CreateDelayedTrigger{
		Trigger: game.DelayedTriggerDef{
			Timing:         timing,
			CapturedObject: opt.Val(eventRef),
			Content:        content,
		},
	}}}}.Ability(), true
}
