package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerControlSpellSequence lowers an ordered effect sequence whose first
// effect (or second, after an initial Untap) is EffectGainControl.  It handles
// two oracle text patterns atomically:
//
//	Pattern A (effects[0] = GainControl):
//	  "Gain control of target X until end of turn. [Untap that X.] [It gains KW.] [Scry N.]"
//
//	Pattern B (effects[0] = Untap, effects[1] = GainControl, same sentence):
//	  "Untap target X and gain control of it until end of turn. [That X gains KW.]"
//
// Subsequent effects (Untap back-ref, keyword grant, counter placement, or
// standalone effects like Scry) are consumed in order.
func lowerControlSpellSequence(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported gain-control spell",
			"the executable source backend supports only exact gain-control sequences targeting one permanent",
		)
	}

	if len(ctx.content.Conditions) != 0 || len(ctx.content.Modes) != 0 {
		return unsupported()
	}
	if len(ctx.content.Targets) != 1 {
		return unsupported()
	}

	// Detect Pattern B: Untap first, GainControl second (same sentence span).
	isPatternB := len(ctx.content.Effects) >= 2 &&
		ctx.content.Effects[0].Kind == compiler.EffectUntap &&
		ctx.content.Effects[1].Kind == compiler.EffectGainControl

	gainControlIdx := 0
	if isPatternB {
		gainControlIdx = 1
	}
	controlEffect := ctx.content.Effects[gainControlIdx]
	if !controlEffect.Exact || controlEffect.Negated {
		return unsupported()
	}
	if isPatternB && (!ctx.content.Effects[0].Exact || ctx.content.Effects[0].Negated) {
		return unsupported()
	}

	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return unsupported()
	}
	// Gaining control of something you already control is a no-op for the
	// control layer, but we do allow ControllerAny (e.g. Threaten) so the
	// effect can still untap and grant keywords.
	if ctx.content.Targets[0].Selector.Controller == compiler.ControllerYou {
		return unsupported()
	}

	var duration game.EffectDuration
	switch controlEffect.Duration {
	case compiler.DurationUntilEndOfTurn:
		duration = game.DurationUntilEndOfTurn
	case compiler.DurationNone:
		duration = game.DurationPermanent
	case compiler.DurationForAsLongAsSourceOnBattlefield:
		duration = game.DurationForAsLongAsSourceOnBattlefield
	case compiler.DurationForAsLongAsYouControlSource:
		duration = game.DurationForAsLongAsYouControlSource
	case compiler.DurationForAsLongAsControlledCreatureEnchanted:
		duration = game.DurationForAsLongAsControlledCreatureEnchanted
	default:
		return unsupported()
	}

	gainControlPrim := game.ApplyContinuous{
		Object: opt.Val(game.TargetPermanentReference(0)),
		ContinuousEffects: []game.ContinuousEffect{{
			Layer:         game.LayerControl,
			NewController: opt.Val(game.Player1),
		}},
		Duration: duration,
	}

	clauseSyntaxes := splitEffectSyntaxes(syntax, ctx.content.Effects)
	consumedTargets := 0
	// Use span-keyed sets to count each reference/keyword exactly once, even
	// when multiple same-sentence effects share the same reference spans.
	consumedRefSpans := make(map[shared.Span]bool)
	consumedKwSpans := make(map[shared.Span]bool)
	var sequence []game.Instruction

	if isPatternB {
		// Pattern B: effects[0] (Untap) and effects[1] (GainControl) share the
		// same sentence span.  Count targets and references from the shared span
		// once rather than per-effect to avoid double-counting.
		sharedSpan := ctx.content.Effects[0].Span
		consumedTargets += len(targetsWithinSpan(ctx.content.Targets, sharedSpan))
		for _, r := range referencesWithinSpan(ctx.content.References, sharedSpan) {
			consumedRefSpans[r.Span] = true
		}
		sequence = append(sequence,
			game.Instruction{Primitive: game.Untap{Object: game.TargetPermanentReference(0)}},
			game.Instruction{Primitive: gainControlPrim},
		)
		for i := 2; i < len(ctx.content.Effects); i++ {
			effAbility := contextForEffect(ctx, &ctx.content.Effects[i])
			prim, ok := lowerControlSequenceFollowOn(cardName, effAbility, &clauseSyntaxes[i])
			if !ok {
				return unsupported()
			}
			sequence = append(sequence, game.Instruction{Primitive: prim})
			for _, r := range effAbility.content.References {
				consumedRefSpans[r.Span] = true
			}
			for _, k := range effAbility.content.Keywords {
				consumedKwSpans[k.Span] = true
			}
		}
	} else {
		// Pattern A: effects[0] is GainControl; subsequent effects are follow-ons.
		effAbility0 := contextForEffect(ctx, &ctx.content.Effects[0])
		consumedTargets += len(effAbility0.content.Targets)
		sequence = append(sequence, game.Instruction{Primitive: gainControlPrim})
		for i := 1; i < len(ctx.content.Effects); i++ {
			effAbility := contextForEffect(ctx, &ctx.content.Effects[i])
			prim, ok := lowerControlSequenceFollowOn(cardName, effAbility, &clauseSyntaxes[i])
			if !ok {
				return unsupported()
			}
			sequence = append(sequence, game.Instruction{Primitive: prim})
			for _, r := range effAbility.content.References {
				consumedRefSpans[r.Span] = true
			}
			for _, k := range effAbility.content.Keywords {
				consumedKwSpans[k.Span] = true
			}
		}
	}

	if consumedTargets != len(ctx.content.Targets) ||
		len(consumedKwSpans) != len(ctx.content.Keywords) ||
		len(consumedRefSpans) != len(ctx.content.References) {
		return unsupported()
	}

	return game.Mode{Targets: []game.TargetSpec{targetSpec}, Sequence: sequence}.Ability(), nil
}

// lowerControlSequenceFollowOn lowers a single follow-on effect in a
// gain-control sequence: an Untap back-reference, a keyword grant, a counter
// placement, or a standalone effect (e.g. Scry) with no back-references.
func lowerControlSequenceFollowOn(
	cardName string,
	ctx contentCtx,
	clauseSyntax *parser.Ability,
) (game.Primitive, bool) {
	effect := ctx.content.Effects[0]
	if !effect.Exact || effect.Negated {
		return nil, false
	}

	switch effect.Kind {
	case compiler.EffectUntap:
		// Back-reference untap: "Untap that creature." — no new targets.
		if len(ctx.content.Targets) != 0 || !referencesTargetZero(ctx.content.References) {
			return nil, false
		}
		return game.Untap{Object: game.TargetPermanentReference(0)}, true

	case compiler.EffectGain:
		// Keyword grant: "It gains haste until end of turn." — back-ref, no new
		// targets. The keyword may carry its own "it"/"that creature" back
		// reference (referencesTargetZero) or, when it rides a combined
		// power/toughness-and-keyword clause ("It gets +2/+0 and gains haste
		// until end of turn."), inherit the prior clause's subject with no
		// reference of its own (EffectContextPriorSubject). Both forms address
		// the controlled creature in slot 0.
		if len(ctx.content.Targets) != 0 || len(ctx.content.Keywords) == 0 ||
			!controlSequenceBackReferencesTargetZero(effect, ctx.content.References) {
			return nil, false
		}
		if effect.Duration != compiler.DurationUntilEndOfTurn {
			return nil, false
		}
		keywords, ok := mixedStaticKeywords(ctx.content.Keywords)
		if !ok {
			return nil, false
		}
		return game.ApplyContinuous{
			Object: opt.Val(game.TargetPermanentReference(0)),
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:       game.LayerAbility,
				AddKeywords: keywords,
			}},
			Duration: game.DurationUntilEndOfTurn,
		}, true

	case compiler.EffectModifyPT:
		// Power/toughness boost back-reference: "It gets +2/+0 until end of
		// turn." / "It gets +X/+0 until end of turn." — back-ref to the
		// controlled creature, no new targets and no keywords (a same-sentence
		// "and gains haste" rider segments into its own EffectGain clause handled
		// above). Only fixed and spell-X deltas lower; dynamic count-based forms
		// ("+1/+1 for each …") fail closed.
		if len(ctx.content.Targets) != 0 || len(ctx.content.Keywords) != 0 ||
			!referencesTargetZero(ctx.content.References) {
			return nil, false
		}
		if effect.Duration != compiler.DurationUntilEndOfTurn {
			return nil, false
		}
		if effect.Amount.DynamicKind != compiler.DynamicAmountNone ||
			!modifyPTSideResolved(effect.PowerDelta) ||
			!modifyPTSideResolved(effect.ToughnessDelta) {
			return nil, false
		}
		return game.ModifyPT{
			Object:         game.TargetPermanentReference(0),
			PowerDelta:     modifyPTSideQuantity(effect.PowerDelta),
			ToughnessDelta: modifyPTSideQuantity(effect.ToughnessDelta),
			Duration:       game.DurationUntilEndOfTurn,
		}, true

	case compiler.EffectPut:
		// Counter placement: "Put a +1/+1 counter on it." — back-ref, no new targets.
		if len(ctx.content.Targets) != 0 || !referencesTargetZero(ctx.content.References) {
			return nil, false
		}
		if !effect.CounterKindKnown || !compiler.CounterKindPlacementSupported(effect.CounterKind) {
			return nil, false
		}
		if !effect.Amount.Known || effect.Amount.Value < 1 {
			return nil, false
		}
		return game.AddCounter{
			Amount:      game.Fixed(effect.Amount.Value),
			Object:      game.TargetPermanentReference(0),
			CounterKind: effect.CounterKind,
		}, true

	default:
		// Standalone effect with no back-references (e.g. Scry).
		if len(ctx.content.References) != 0 || len(ctx.content.Targets) != 0 {
			return nil, false
		}
		content, diag := lowerSingleEffectSpell(cardName, ctx, clauseSyntax)
		if diag != nil {
			return nil, false
		}
		if len(content.SharedTargets) != 0 ||
			content.IsModal() ||
			len(content.Modes) != 1 ||
			len(content.Modes[0].Targets) != 0 ||
			len(content.Modes[0].Sequence) != 1 {
			return nil, false
		}
		return content.Modes[0].Sequence[0].Primitive, true
	}
}

func referencesTargetZero(references []compiler.CompiledReference) bool {
	return len(references) == 1 &&
		references[0].Binding == compiler.ReferenceBindingTarget &&
		references[0].Occurrence == 0
}

// controlSequenceBackReferencesTargetZero reports whether a gain-control
// sequence follow-on addresses the controlled creature in target slot 0. The
// follow-on either carries its own "it"/"that creature" reference to that slot
// (referencesTargetZero) or, when it is the trailing keyword half of a combined
// power/toughness-and-keyword clause, inherits the prior clause's subject with
// no reference of its own (EffectContextPriorSubject). Within a gain-control
// sequence every prior clause already addresses slot 0, so the inherited
// subject is that same creature.
func controlSequenceBackReferencesTargetZero(
	effect compiler.CompiledEffect,
	references []compiler.CompiledReference,
) bool {
	if referencesTargetZero(references) {
		return true
	}
	return effect.Context == parser.EffectContextPriorSubject && len(references) == 0
}

// referencesSourceSelfOnly reports whether every reference (if any) binds to the
// source permanent. Self-relative gain-control durations restate the source
// ("this creature remains on the battlefield") as a back-reference that carries
// no independent action, so the gain-control lowering accepts those references
// without a corresponding instruction.
func referencesSourceSelfOnly(references []compiler.CompiledReference) bool {
	for i := range references {
		if references[i].Binding != compiler.ReferenceBindingSource {
			return false
		}
	}
	return true
}

// lowerSingleControlSpell lowers a single EffectGainControl spell with no
// Untap or keyword grant (e.g. "Gain control of target permanent." or the
// DurationUntilEndOfTurn variant).
func lowerSingleControlSpell(
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported gain-control spell",
			"the executable source backend supports only exact gain-control of one target permanent",
		)
	}
	if len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!ctx.content.Effects[0].Exact ||
		ctx.content.Effects[0].Negated {
		return unsupported()
	}
	if ctx.content.Effects[0].Context == parser.EffectContextTarget {
		return lowerGiveControlSpell(ctx)
	}
	if len(ctx.content.Targets) != 1 {
		return unsupported()
	}
	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return unsupported()
	}
	if ctx.content.Targets[0].Selector.Controller == compiler.ControllerYou {
		return unsupported()
	}
	var duration game.EffectDuration
	switch ctx.content.Effects[0].Duration {
	case compiler.DurationUntilEndOfTurn:
		duration = game.DurationUntilEndOfTurn
	case compiler.DurationNone:
		duration = game.DurationPermanent
	case compiler.DurationForAsLongAsSourceOnBattlefield:
		duration = game.DurationForAsLongAsSourceOnBattlefield
	case compiler.DurationForAsLongAsYouControlSource:
		duration = game.DurationForAsLongAsYouControlSource
	case compiler.DurationForAsLongAsControlledCreatureEnchanted:
		duration = game.DurationForAsLongAsControlledCreatureEnchanted
	case compiler.DurationForAsLongAsThatPlayerIsMonarch:
		duration = game.DurationForAsLongAsPlayerIsMonarch
	default:
		return unsupported()
	}
	// Self-relative durations ("for as long as this creature remains on the
	// battlefield" / "for as long as you control this creature" / "for as long
	// as that creature is enchanted") may carry a single back-reference to the
	// source or the controlled creature. The duration enum already encodes the
	// boundary, so the reference needs no lowering action of its own; accept it
	// for these durations only. The monarch duration ("for as long as they're
	// the monarch") likewise carries the "that player" reference the target's
	// controlled-by-event-player restriction already consumes. Every other
	// duration requires no references.
	switch duration {
	case game.DurationForAsLongAsControlledCreatureEnchanted:
		if len(ctx.content.References) != 0 && !referencesTargetZero(ctx.content.References) {
			return unsupported()
		}
	case game.DurationForAsLongAsPlayerIsMonarch:
		if !referencesTargetZero(ctx.content.References) {
			return unsupported()
		}
	case game.DurationForAsLongAsSourceOnBattlefield, game.DurationForAsLongAsYouControlSource:
		if !referencesSourceSelfOnly(ctx.content.References) {
			return unsupported()
		}
	default:
		if len(ctx.content.References) != 0 {
			return unsupported()
		}
	}
	continuousEffect := game.ContinuousEffect{
		Layer:         game.LayerControl,
		NewController: opt.Val(game.Player1),
	}
	// The monarch duration binds its expiry to the player who became the
	// monarch — the triggering event player — resolved when the continuous
	// effect is created.
	if duration == game.DurationForAsLongAsPlayerIsMonarch {
		continuousEffect.ExpiresForRef = opt.Val(game.EventPlayerReference())
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				Object:            opt.Val(game.TargetPermanentReference(0)),
				ContinuousEffects: []game.ContinuousEffect{continuousEffect},
				Duration:          duration,
			},
		}},
	}.Ability(), nil
}

// lowerGiveControlSpell lowers the give-control forms whose subject is a target
// player who gains control of a permanent (EffectContextTarget). The new
// controller is the chosen target player, resolved at application time through
// NewControllerRef. Two shapes are supported:
//
//	Two-target: "Target player gains control of target permanent you control."
//	  (Donate, Harmless Offering, Wrong Turn) — target slot 0 is the player and
//	  target slot 1 is the controlled permanent.
//
//	Source self-gift: "Target opponent gains control of this <object>."
//	  (Jinxed Idol, Avarice Amulet, Measure of Wickedness) — target slot 0 is the
//	  player and the controlled object is the ability's own source.
//
// Unlike the controller-subject gain-control forms, a "you control" object is
// the whole point of giving a permanent away, so it is accepted here.
func lowerGiveControlSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported gain-control spell",
			"the executable source backend supports only exact give-control to a target player",
		)
	}
	if len(ctx.content.Targets) == 0 {
		return unsupported()
	}
	playerSpec, ok := playerTargetSpec(ctx.content.Targets[0])
	if !ok {
		return unsupported()
	}

	var duration game.EffectDuration
	switch ctx.content.Effects[0].Duration {
	case compiler.DurationUntilEndOfTurn:
		duration = game.DurationUntilEndOfTurn
	case compiler.DurationNone:
		duration = game.DurationPermanent
	default:
		return unsupported()
	}

	var object game.ObjectReference
	targets := []game.TargetSpec{playerSpec}
	switch {
	case len(ctx.content.Targets) == 2 && len(ctx.content.References) == 0:
		permSpec, ok := permanentTargetSpec(ctx.content.Targets[1])
		if !ok {
			return unsupported()
		}
		object = game.TargetPermanentReference(1)
		targets = append(targets, permSpec)
	case len(ctx.content.Targets) == 1 && referencesSourceSelfOnly(ctx.content.References):
		if len(ctx.content.References) != 1 {
			return unsupported()
		}
		object = game.SourcePermanentReference()
	default:
		return unsupported()
	}

	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				Object: opt.Val(object),
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:            game.LayerControl,
					NewControllerRef: opt.Val(game.TargetPlayerReference(0)),
				}},
				Duration: duration,
			},
		}},
	}.Ability(), nil
}
