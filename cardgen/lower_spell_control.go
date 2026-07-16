package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerGiveControlToEventPlayerSpell lowers the give-control form whose recipient
// is the triggering event's player and whose object is the ability's own source
// permanent, for the turn ("you may have that player gain control of Slicer until
// end of turn", Slicer, Hired Muscle, gated by the sequence's optional flow). The
// new controller is the player who triggered the ability — the opponent whose
// upkeep it is — resolved at application time through
// NewControllerRef = EventPlayerReference(); the given permanent is the source.
// It fails closed unless the effect is exactly the source-object, event-player-
// recipient, until-end-of-turn shape carrying no targets, keywords, conditions,
// or modes and exactly the source and event-player references.
func lowerGiveControlToEventPlayerSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				Object: opt.Val(game.SourcePermanentReference()),
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:            game.LayerControl,
					NewControllerRef: opt.Val(game.EventPlayerReference()),
				}},
				Duration: game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability(), nil
}

// givesControlOfSourceToEventPlayer reports whether the content is exactly "have
// that player gain control of <source> until end of turn": a single
// until-end-of-turn gain-control effect resolved in the triggering player's
// context, referencing only the ability's own source (the given object) and the
// event player (the control recipient), with no targets, keywords, conditions, or
// modes layered on. Only this exact shape is lowered by
// lowerGiveControlToEventPlayerSpell; any other referenced-player control text
// falls through to the ordinary unsupported gain-control diagnostic.
func givesControlOfSourceToEventPlayer(ctx contentCtx) bool {
	effect := ctx.content.Effects[0]
	if effect.Context != parser.EffectContextReferencedPlayer {
		return false
	}
	if effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationUntilEndOfTurn ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 {
		return false
	}
	return referencesSourceAndEventPlayer(ctx.content.References)
}

// referencesSourceAndEventPlayer reports whether references are exactly the
// ability's own source permanent (the given object, "Slicer") and the triggering
// event's player (the control recipient, "that player"). Both are consumed by the
// give-control-to-event-player lowering: the source binds the given object and
// the event player binds NewControllerRef, so no reference is left dropped.
func referencesSourceAndEventPlayer(references []compiler.CompiledReference) bool {
	if len(references) != 2 {
		return false
	}
	source, eventPlayer := 0, 0
	for _, reference := range references {
		switch reference.Binding {
		case compiler.ReferenceBindingSource:
			source++
		case compiler.ReferenceBindingEventPlayer:
			eventPlayer++
		default:
			return false
		}
	}
	return source == 1 && eventPlayer == 1
}

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

	if len(ctx.content.Modes) != 0 {
		return unsupported()
	}
	if len(ctx.content.Targets) != 1 {
		return unsupported()
	}

	// A gain-control sequence may carry a supported ability-level condition that
	// gates a follow-on effect ("If X is 5 or more, create a token that's a copy
	// of that creature."). Match each condition to the follow-on effect whose
	// clause span contains it and lower it as an effect gate through the shared
	// per-effect condition path, exactly as lowerOrderedEffectSequence does. The
	// gain-control effect itself is never gated: a condition matched to it (or,
	// in Pattern B, to the leading untap) is left unapplied and fails the
	// consumed-condition count check below, so the sequence stays fail-closed.
	effectConditions, _, ok := matchSequenceEffectConditions(ctx.content.Effects, ctx.content.Conditions)
	if !ok {
		return unsupported()
	}
	consumedConditions := 0

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
			instruction := game.Instruction{Primitive: prim}
			if gate, gated := effectConditions[i]; gated {
				instruction.Condition = opt.Val(gate)
				consumedConditions++
			}
			sequence = append(sequence, instruction)
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
			instruction := game.Instruction{Primitive: prim}
			if gate, gated := effectConditions[i]; gated {
				instruction.Condition = opt.Val(gate)
				consumedConditions++
			}
			sequence = append(sequence, instruction)
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
		len(consumedRefSpans) != len(ctx.content.References) ||
		consumedConditions != len(effectConditions) {
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
		// Copy-token back-reference: "create a token that's a copy of that
		// creature." The token copies the gain-control target in slot 0; the
		// leading "that creature" reference binds that target. Delegate to the
		// shared copy-token lowerer — the same path Yenna and Molten Duplication
		// use for "a copy of it/target <permanent>" — and clear the ability-level
		// condition inherited from the parent gain-control context (the sequence
		// applies it as an effect gate on the produced instruction). A standalone
		// token creation ("Create a Blood token.") carries no reference and falls
		// through unchanged to the standalone handling below.
		if effect.Kind == compiler.EffectCreate &&
			len(ctx.content.Targets) == 0 &&
			len(ctx.content.References) != 0 &&
			referencesBindTo(ctx.content.References[:1], compiler.ReferenceBindingTarget, 0) {
			createCtx := ctx
			createCtx.content.Conditions = nil
			content, diag := lowerCreateTokenSpellLinked(createCtx, "")
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
	if givesControlOfSourceToEventPlayer(ctx) {
		return lowerGiveControlToEventPlayerSpell(ctx)
	}
	// A mass gain-control clause ("gain control of all <group> [that a player
	// controls]") selects a group rather than a single permanent. It is
	// recognized by the effect's plural "all" selector and lowered to a
	// LayerControl continuous effect over a GroupReference, distinct from the
	// single-target path below.
	if ctx.content.Effects[0].Selector.All {
		return lowerMassControlSpell(ctx)
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

// lowerMassControlSpell lowers the mass/group gain-control forms whose subject is
// a group of permanents ("gain control of all <filtered permanents> [that a
// player controls]") rather than a single target. The resolving ability's
// controller gains control of every permanent the group matches, modeled as a
// LayerControl continuous effect carried on a GroupReference instead of a
// per-card loop or name list.
//
// Because the effect is generated by the resolution of a spell or ability, the
// runtime locks the affected set of permanents when the effect begins (CR
// 611.2c): applyTypedContinuousEffects snapshots the group's members at
// resolution into one fixed-object control effect each, so permanents that later
// come to satisfy the group relationship are not taken and the source permanent
// leaving does not revert a permanent-duration effect. This matches Hellkite
// Tyrant, whose combat-damage trigger permanently steals exactly the artifacts
// the damaged player controlled as the ability resolved.
//
// The new controller is the resolving controller, carried as the Player1 sentinel
// the runtime substitutes before it expands the group. Three group anchors are
// supported, mirroring the existing mass tap/untap/counter machinery:
//
//   - an explicit selector controller or none ("gain control of all creatures",
//     "gain control of all enchantments", Aura Thief) via exactMassGroup;
//   - the triggering event's player ("gain control of all artifacts that player
//     controls", Hellkite Tyrant) via eventPlayerControlledMassGroup; and
//   - a targeted player or opponent ("gain control of all creatures target
//     opponent controls", Ashiok) via targetPlayerControlledMassControlGroup.
//
// Only the permanent (default) and until-end-of-turn durations lower; every other
// duration, any condition, mode, keyword, or optional offer, a negated or
// non-controller effect, and any group the projectors cannot express fail closed.
func lowerMassControlSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported gain-control spell",
			"the executable source backend supports only mass gain-control of a selector or player-controlled group",
		)
	}
	effect := ctx.content.Effects[0]
	if effect.Negated ||
		effect.Optional ||
		ctx.optional ||
		effect.Context != parser.EffectContextController ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Keywords) != 0 {
		return unsupported()
	}
	var duration game.EffectDuration
	switch effect.Duration {
	case compiler.DurationNone:
		duration = game.DurationPermanent
	case compiler.DurationUntilEndOfTurn:
		duration = game.DurationUntilEndOfTurn
	default:
		return unsupported()
	}
	group, targets, ok := massControlGroup(ctx)
	if !ok {
		return unsupported()
	}
	instruction := continuousGroupInstruction(
		group,
		[]game.ContinuousEffect{{
			Layer:         game.LayerControl,
			NewController: opt.Val(game.Player1),
		}},
		duration,
	)
	return game.Mode{
		Targets:  targets,
		Sequence: []game.Instruction{instruction},
	}.Ability(), nil
}

// massControlGroup resolves the group a mass gain-control clause selects, along
// with any player target its controller anchor introduces. It tries, in order:
// the selector-controller mass group (exactMassGroup: "all creatures", "all
// artifacts you control"), which carries no target; the triggering event player's
// controlled group (eventPlayerControlledMassGroup: "all artifacts that player
// controls"), which carries no target; and a targeted player's controlled group
// (targetPlayerControlledMassControlGroup: "all creatures target opponent
// controls"), which carries the single player target. It reports false when no
// projector matches, so lowerMassControlSpell fails closed.
func massControlGroup(ctx contentCtx) (game.GroupReference, []game.TargetSpec, bool) {
	if group, ok := exactMassGroup(ctx); ok {
		return group, nil, true
	}
	if group, ok := eventPlayerControlledMassGroup(ctx); ok {
		return group, nil, true
	}
	return targetPlayerControlledMassControlGroup(ctx)
}

// targetPlayerControlledMassControlGroup resolves the group of a mass gain-control
// whose controller anchor is a single targeted player or opponent ("gain control
// of all creatures target opponent controls", Ashiok, Dream Render). The targeted
// player supplies the group's controller relationship through
// TargetPlayerReference(0); the effect's plural selector, which must carry no
// explicit controller of its own, supplies the group's type filter. The target is
// reconstructed by targetControlsPlayerSpec (the trailing "controls" relationship
// is rebuilt from the selector kind), mirroring the target-player counter
// placement path so the two mass families stay consistent.
//
// It fails closed for any additional target or reference, a selector carrying its
// own controller, a target the player-target spec cannot express, or a group the
// mass selection projection cannot express.
func targetPlayerControlledMassControlGroup(
	ctx contentCtx,
) (game.GroupReference, []game.TargetSpec, bool) {
	if len(ctx.content.Targets) != 1 || len(ctx.content.References) != 0 {
		return game.GroupReference{}, nil, false
	}
	effect := ctx.content.Effects[0]
	if !effect.Selector.All || effect.Selector.Controller != compiler.ControllerAny {
		return game.GroupReference{}, nil, false
	}
	target, ok := targetControlsPlayerSpec(ctx.content.Targets[0])
	if !ok {
		return game.GroupReference{}, nil, false
	}
	selection, ok := massGroupSelection(effect.Selector)
	if !ok {
		return game.GroupReference{}, nil, false
	}
	group := game.PlayerControlledGroup(game.TargetPlayerReference(0), selection)
	if len(group.Validate()) != 0 {
		return game.GroupReference{}, nil, false
	}
	return group, []game.TargetSpec{target}, true
}
