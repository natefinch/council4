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
		// Keyword grant: "It gains haste until end of turn." — back-ref, no new targets.
		if len(ctx.content.Targets) != 0 || len(ctx.content.Keywords) == 0 ||
			!referencesTargetZero(ctx.content.References) {
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
	if len(ctx.content.Targets) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!ctx.content.Effects[0].Exact ||
		ctx.content.Effects[0].Negated {
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
	default:
		return unsupported()
	}
	// The "for as long as that creature is enchanted" wording carries a single
	// back-reference to the controlled creature (the lone target). It needs no
	// lowering action of its own, so accept it for that duration only; every
	// other supported duration requires no references.
	if duration == game.DurationForAsLongAsControlledCreatureEnchanted {
		if len(ctx.content.References) != 0 && !referencesTargetZero(ctx.content.References) {
			return unsupported()
		}
	} else if len(ctx.content.References) != 0 {
		return unsupported()
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				Object: opt.Val(game.TargetPermanentReference(0)),
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:         game.LayerControl,
					NewController: opt.Val(game.Player1),
				}},
				Duration: duration,
			},
		}},
	}.Ability(), nil
}
