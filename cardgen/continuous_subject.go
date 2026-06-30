package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// This file holds the shared composition helpers that route a set of
// game.ContinuousEffect values to whichever subject a one-shot continuous effect
// selects — the source permanent, a single permanent target, a referenced
// object, or a static creature/permanent group — and assemble the resulting
// ApplyContinuous mode. Continuous-effect lowerers (double power/toughness, set
// base power/toughness, color change, self/target animation, keyword grant/loss,
// power/toughness switch) build their typed continuous effects independently and
// then delegate subject selection here so the same recipient plumbing is not
// re-derived per effect family.

// continuousSubjectOptions configures continuousSubjectMode: which subjects a set
// of continuous effects may be routed to. The referenced-object subject is
// intentionally excluded; its recognition gating (which reference bindings and
// effect contexts are valid) is effect-specific, so callers resolve a reference
// themselves and assemble it with continuousObjectMode.
type continuousSubjectOptions struct {
	// SourceForm reports that the typed effect selects the source permanent
	// through its own boolean signal (such as EffectModifyPT's SetBasePTSource or
	// EffectBecomeColor's BecomeColorSource) rather than through a reference. When
	// set, the content must carry no targets and no references.
	SourceForm bool
	// AllowGroup permits a static-subject creature/permanent group subject.
	AllowGroup bool
	// AllowTarget permits a single permanent target subject.
	AllowTarget bool
}

// continuousSubjectMode routes continuousEffects to the subject a one-shot
// continuous effect selects and builds the ApplyContinuous mode for the given
// duration. It handles the three subjects whose recognition is uniform across
// effect families: the source permanent (opts.SourceForm), a static
// creature/permanent group (opts.AllowGroup), and a single permanent target
// (opts.AllowTarget). Referenced-object subjects are resolved by callers before
// delegating here, because their valid bindings differ per effect. Any shape the
// router cannot reduce, or a subject the caller disallows, fails closed via
// unsupported.
func continuousSubjectMode(
	ctx contentCtx,
	effect *compiler.CompiledEffect,
	continuousEffects []game.ContinuousEffect,
	duration game.EffectDuration,
	opts continuousSubjectOptions,
	unsupported func() (game.AbilityContent, *shared.Diagnostic),
) (game.AbilityContent, *shared.Diagnostic) {
	if opts.SourceForm {
		if len(ctx.content.Targets) != 0 || len(ctx.content.References) != 0 {
			return unsupported()
		}
		return continuousSourceMode(continuousEffects, duration), nil
	}
	if effect.StaticSubject != compiler.StaticSubjectNone {
		if !opts.AllowGroup || len(ctx.content.Targets) != 0 || len(ctx.content.References) != 0 {
			return unsupported()
		}
		group, ok := resolvingStaticSubjectGroup(effect)
		if !ok {
			return unsupported()
		}
		return continuousGroupMode(group, continuousEffects, duration), nil
	}
	if opts.AllowTarget && len(ctx.content.Targets) == 1 && len(ctx.content.References) == 0 {
		return continuousTargetMode(ctx.content.Targets[0], continuousEffects, duration, unsupported)
	}
	return unsupported()
}

// continuousSourceMode builds an ApplyContinuous mode applying the given
// continuous effects to the source permanent for the given duration.
func continuousSourceMode(continuousEffects []game.ContinuousEffect, duration game.EffectDuration) game.AbilityContent {
	return continuousObjectMode(game.SourcePermanentReference(), continuousEffects, duration)
}

// continuousObjectMode builds an ApplyContinuous mode applying the given
// continuous effects to a single resolved object for the given duration.
func continuousObjectMode(
	object game.ObjectReference,
	continuousEffects []game.ContinuousEffect,
	duration game.EffectDuration,
) game.AbilityContent {
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				Object:            opt.Val(object),
				ContinuousEffects: continuousEffects,
				Duration:          duration,
			},
		}},
	}.Ability()
}

// continuousGroupMode builds an ApplyContinuous mode applying the given
// continuous effects to a never-resolving static group for the given duration.
// The group reference rides on each continuous effect, so the effects are cloned
// before the group is stamped to leave the caller's slice untouched.
func continuousGroupMode(
	group game.GroupReference,
	continuousEffects []game.ContinuousEffect,
	duration game.EffectDuration,
) game.AbilityContent {
	grouped := slices.Clone(continuousEffects)
	for i := range grouped {
		grouped[i].Group = group
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				ContinuousEffects: grouped,
				Duration:          duration,
			},
		}},
	}.Ability()
}

// temporaryKeywordTargetMode builds the until-end-of-turn ApplyContinuous mode
// for a keyword grant or loss applied to one permanent target slot per chosen
// target. The target may be single ("Target creature gains flying…"), optional
// ("up to one target creature gains…"), or multi-cardinality; a declined "up to"
// slot leaves an unresolved target index the runtime ApplyContinuous no-ops, so
// only chosen permanents are affected. The target's filter is validated by the
// canonical permanentTargetSpecWithCardinality, so the same subtype, card-type,
// color, and tapped restrictions destroy and exile already target (e.g. "target
// Human", "target artifact", "target black creature") apply here too. It fails
// closed for any target permanentTargetSpecWithCardinality cannot express.
func temporaryKeywordTargetMode(
	target compiler.CompiledTarget,
	continuousEffects []game.ContinuousEffect,
	unsupported func() (game.AbilityContent, *shared.Diagnostic),
) (game.AbilityContent, *shared.Diagnostic) {
	return continuousTargetMode(target, continuousEffects, game.DurationUntilEndOfTurn, unsupported)
}

// continuousTargetMode builds an ApplyContinuous mode that applies the given
// continuous effects to each targeted permanent for the given duration. It backs
// both the until-end-of-turn keyword/polymorph forms and the permanent
// named-become polymorph. It fails closed when the target cannot reduce to a
// permanent target spec.
func continuousTargetMode(
	target compiler.CompiledTarget,
	continuousEffects []game.ContinuousEffect,
	duration game.EffectDuration,
	unsupported func() (game.AbilityContent, *shared.Diagnostic),
) (game.AbilityContent, *shared.Diagnostic) {
	spec, ok := permanentTargetSpecWithCardinality(target)
	if !ok || spec.MaxTargets < 1 {
		return unsupported()
	}
	sequence := make([]game.Instruction, 0, spec.MaxTargets)
	for i := range spec.MaxTargets {
		sequence = append(sequence, game.Instruction{
			Primitive: game.ApplyContinuous{
				Object:            opt.Val(game.TargetPermanentReference(i)),
				ContinuousEffects: continuousEffects,
				Duration:          duration,
			},
		})
	}
	return game.Mode{
		Targets:  []game.TargetSpec{spec},
		Sequence: sequence,
	}.Ability(), nil
}
