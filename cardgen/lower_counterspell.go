package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const delayedDrawAmountChoiceKey = game.ChoiceKey("delayed-draw-amount")

func lowerCounterThenNextMainManaSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 1 {
		return game.AbilityContent{}, false
	}
	counterEffect := &ctx.content.Effects[0]
	manaEffect := &ctx.content.Effects[1]
	targetSpec, ok := stackSpellTargetSpec(ctx.content.Targets[0])
	if !ok ||
		!isExactMandatoryCounterEffect(counterEffect, ctx.content.Targets[0]) ||
		counterEffect.Connection != parser.EffectConnectionNone ||
		manaEffect.Kind != compiler.EffectAddMana ||
		!manaEffect.Exact ||
		manaEffect.Negated ||
		manaEffect.Optional ||
		manaEffect.Connection != parser.EffectConnectionNone ||
		manaEffect.Context != parser.EffectContextController ||
		manaEffect.DelayedTiming != game.DelayedAtBeginningOfNextMainPhase ||
		manaEffect.Duration != compiler.DurationNone ||
		!manaEffect.Mana.DynamicColorless ||
		manaEffect.Amount.DynamicKind != compiler.DynamicAmountSourceManaValue ||
		manaEffect.Amount.DynamicForm != compiler.DynamicAmountEqual ||
		manaEffect.Amount.Multiplier != 1 ||
		len(manaEffect.Targets) != 0 ||
		!referencesBindTo(manaEffect.References, compiler.ReferenceBindingTarget, 0) {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)}},
			{Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
				Timing: game.DelayedAtBeginningOfNextMainPhase,
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.AddMana{
						Amount: game.Dynamic(game.DynamicAmount{
							Kind:   game.DynamicAmountCapturedTargetManaValue,
							Object: game.CapturedTargetStackObjectReference(0),
						}),
						ManaColor: mana.C,
					},
				}}}.Ability(),
			}}},
		},
	}.Ability(), true
}

func lowerCounterThenNextTurnUpkeepDrawAbilities(cardName string, compilation compiler.Compilation) (game.AbilityContent, bool) {
	if len(compilation.Abilities) < 2 ||
		len(compilation.Abilities) != len(compilation.Syntax.Abilities) {
		return game.AbilityContent{}, false
	}
	content := compiler.AbilityContent{
		Span: shared.Span{
			Start: compilation.Abilities[0].Span.Start,
			End:   compilation.Abilities[len(compilation.Abilities)-1].Span.End,
		},
	}
	for _, ability := range compilation.Abilities {
		if ability.Kind != compiler.AbilitySpell ||
			ability.Optional ||
			ability.Cost != nil ||
			ability.Trigger != nil ||
			ability.Static != nil {
			return game.AbilityContent{}, false
		}
		content.Modes = append(content.Modes, ability.Content.Modes...)
		content.Targets = append(content.Targets, ability.Content.Targets...)
		content.Conditions = append(content.Conditions, ability.Content.Conditions...)
		content.Effects = append(content.Effects, ability.Content.Effects...)
		content.Keywords = append(content.Keywords, ability.Content.Keywords...)
		content.References = append(content.References, ability.Content.References...)
	}
	result, ok := lowerCounterThenNextTurnUpkeepDraws(contentCtx{
		span:    content.Span,
		content: content,
	})
	if !ok {
		return game.AbilityContent{}, false
	}
	for i, ability := range compilation.Abilities {
		check := ability
		check.Content.Modes = append([]compiler.CompiledMode(nil), ability.Content.Modes...)
		check.Content.Targets = append([]compiler.CompiledTarget(nil), ability.Content.Targets...)
		check.Content.Conditions = append([]compiler.CompiledCondition(nil), ability.Content.Conditions...)
		check.Content.Effects = append([]compiler.CompiledEffect(nil), ability.Content.Effects...)
		check.Content.Keywords = append([]compiler.CompiledKeyword(nil), ability.Content.Keywords...)
		check.Content.References = append([]compiler.CompiledReference(nil), ability.Content.References...)
		lowered, diagnostic := lowerExecutableAbility(
			cardName,
			false,
			nil,
			check,
			&compilation.Syntax.Abilities[i],
		)
		if diagnostic != nil || !lowered.complete(check, &compilation.Syntax.Abilities[i]) {
			return game.AbilityContent{}, false
		}
	}
	return result, true
}

func lowerCounterThenNextTurnUpkeepDraws(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) < 2 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	counterEffect := ctx.content.Effects[0]
	if counterEffect.Kind != compiler.EffectCounter ||
		!counterEffect.Exact ||
		counterEffect.Negated ||
		counterEffect.Optional ||
		counterEffect.Context != parser.EffectContextController ||
		counterEffect.DelayedTiming != 0 ||
		counterEffect.Duration != compiler.DurationNone ||
		counterEffect.Amount.Known ||
		counterEffect.Amount.RangeKnown ||
		len(counterEffect.Targets) != 1 ||
		len(counterEffect.References) != 0 {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := stackSpellTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	sequence := []game.Instruction{{
		Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)},
	}}
	referenceCount := 0
	for i := 1; i < len(ctx.content.Effects); i++ {
		effect := ctx.content.Effects[i]
		delayed, refs, ok := lowerNextTurnUpkeepDraw(&effect)
		if !ok {
			return game.AbilityContent{}, false
		}
		referenceCount += refs
		sequence = append(sequence, game.Instruction{Primitive: delayed})
	}
	if referenceCount != len(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	consumed := ctx
	consumed.content.Targets = nil
	consumed.content.References = nil
	if consumed.content.Unconsumed() {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets:  []game.TargetSpec{targetSpec},
		Sequence: sequence,
	}.Ability(), true
}

func lowerNextTurnUpkeepDraw(effect *compiler.CompiledEffect) (game.CreateDelayedTrigger, int, bool) {
	if effect.Kind != compiler.EffectDraw ||
		!effect.Exact ||
		effect.Negated ||
		effect.DelayedTiming != game.DelayedAtBeginningOfNextUpkeep ||
		effect.Duration != compiler.DurationNone ||
		len(effect.Targets) != 0 {
		return game.CreateDelayedTrigger{}, 0, false
	}
	trigger := game.DelayedTriggerDef{Timing: game.DelayedAtBeginningOfNextUpkeep}
	drawPlayer := game.ControllerReference()
	var choicePlayer *game.PlayerReference
	referenceCount := 0
	switch effect.Context {
	case parser.EffectContextController:
		if len(effect.References) != 0 {
			return game.CreateDelayedTrigger{}, 0, false
		}
	case parser.EffectContextReferencedObjectController:
		if !referencesBindTo(effect.References, compiler.ReferenceBindingTarget, 0) {
			return game.CreateDelayedTrigger{}, 0, false
		}
		referenceCount = len(effect.References)
		drawPlayer = game.CapturedTargetControllerReference(0)
		choicePlayer = &drawPlayer
	default:
		return game.CreateDelayedTrigger{}, 0, false
	}
	switch {
	case effect.Amount.RangeKnown &&
		effect.Amount.Minimum == 0 &&
		effect.Amount.Maximum > 0:
		trigger.Content = game.Mode{Sequence: []game.Instruction{
			{
				Primitive: game.Choose{
					Choice: game.ResolutionChoice{
						Kind:            game.ResolutionChoiceNumber,
						Prompt:          "Choose how many cards to draw.",
						PlayerReference: choicePlayer,
						MinNumber:       effect.Amount.Minimum,
						MaxNumber:       effect.Amount.Maximum,
					},
					PublishChoice: delayedDrawAmountChoiceKey,
				},
			},
			{
				Primitive: game.Draw{
					Amount: game.Dynamic(game.DynamicAmount{
						Kind:      game.DynamicAmountChosenNumber,
						ResultKey: game.ResultKey(delayedDrawAmountChoiceKey),
					}),
					Player: drawPlayer,
				},
			},
		}}.Ability()
	case effect.Amount.Known && effect.Amount.Value > 0 &&
		(effect.Context == parser.EffectContextController || !effect.Optional):
		trigger.Optional = effect.Optional
		trigger.Content = game.Mode{Sequence: []game.Instruction{{
			Primitive: game.Draw{
				Amount: game.Fixed(effect.Amount.Value),
				Player: drawPlayer,
			},
		}}}.Ability()
	default:
		return game.CreateDelayedTrigger{}, 0, false
	}
	return game.CreateDelayedTrigger{Trigger: trigger}, referenceCount, true
}

func lowerCounterThenTargetControllerTokenSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if !isCounterThenCreateSequence(ctx.content) ||
		!hasExactLinkedCounterTokenEnvelope(ctx) {
		return game.AbilityContent{}, false
	}
	counterEffect := &ctx.content.Effects[0]
	tokenEffect := &ctx.content.Effects[1]
	target := ctx.content.Targets[0]
	if !isExactMandatoryCounterEffect(counterEffect, target) {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := stackSpellTargetSpec(target)
	if !ok {
		return game.AbilityContent{}, false
	}
	if !isTargetControllerTokenEffectForTarget(tokenEffect, 0) {
		return game.AbilityContent{}, false
	}
	tokenInstruction, ok := targetControllerTokenInstruction(ctx, tokenEffect)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)}},
			tokenInstruction,
		},
	}.Ability(), true
}

// targetControllerTokenInstruction builds the CreateToken instruction whose
// recipient is the controller of the countered spell (target stack object 0).
// It accepts the same unmodified creature and predefined-artifact token
// definitions as standalone token creation, plus a fixed, X, or rules-derived
// count, but rejects tapped, attacking, copy, and choice token shapes.
func targetControllerTokenInstruction(ctx contentCtx, tokenEffect *compiler.CompiledEffect) (game.Instruction, bool) {
	if tokenEffect.TokenCopyOfTarget ||
		tokenEffect.TokenChoice ||
		tokenEffect.Selector.Tapped ||
		tokenEffect.Selector.Attacking {
		return game.Instruction{}, false
	}
	def, ok := synthesizeCreatureTokenDef(tokenEffect, nil)
	if !ok {
		def, ok = synthesizeNamedArtifactTokenDef(tokenEffect)
	}
	if !ok {
		return game.Instruction{}, false
	}
	amount, ok := createTokenAmount(ctx, tokenEffect, game.ObjectReference{})
	if !ok {
		return game.Instruction{}, false
	}
	return game.Instruction{Primitive: game.CreateToken{
		Amount:    amount,
		Source:    game.TokenDef(def),
		Recipient: opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
	}}, true
}

func isCounterThenCreateSequence(content compiler.AbilityContent) bool {
	if len(content.Effects) != 2 {
		return false
	}
	counterEffect := content.Effects[0]
	tokenEffect := content.Effects[1]
	return counterEffect.Kind == compiler.EffectCounter &&
		tokenEffect.Kind == compiler.EffectCreate &&
		counterEffect.Connection == parser.EffectConnectionNone &&
		tokenEffect.Connection == parser.EffectConnectionNone
}

func hasExactLinkedCounterTokenEnvelope(ctx contentCtx) bool {
	return !ctx.optional &&
		len(ctx.content.Targets) == 1 &&
		len(ctx.content.Conditions) == 0 &&
		len(ctx.content.Keywords) == 0 &&
		len(ctx.content.Modes) == 0 &&
		len(ctx.content.References) == 1
}

func isExactMandatoryEffect(effect *compiler.CompiledEffect) bool {
	return effect.Exact &&
		!effect.Optional &&
		!effect.Negated &&
		effect.DelayedTiming == 0 &&
		effect.Duration == compiler.DurationNone
}

func isExactMandatoryCounterEffect(
	effect *compiler.CompiledEffect,
	target compiler.CompiledTarget,
) bool {
	return effect.Kind == compiler.EffectCounter &&
		effect.Context == parser.EffectContextController &&
		isExactMandatoryEffect(effect) &&
		!effect.Amount.Known &&
		len(effect.Payment.ManaCost) == 0 &&
		len(effect.Targets) == 1 &&
		target.Exact &&
		effect.Targets[0].Span == target.Span
}

func isTargetControllerTokenEffectForTarget(
	effect *compiler.CompiledEffect,
	targetIndex int,
) bool {
	return effect.Kind == compiler.EffectCreate &&
		effect.Context == parser.EffectContextReferencedObjectController &&
		isExactMandatoryEffect(effect) &&
		len(effect.Targets) == 0 &&
		referencesBindOnlyToTarget(effect.References, targetIndex) &&
		referencesBindOnlyToTarget(effect.SubjectReferences, targetIndex)
}

func referencesBindOnlyToTarget(references []compiler.CompiledReference, targetIndex int) bool {
	return len(references) == 1 &&
		referencesBindTo(references, compiler.ReferenceBindingTarget, targetIndex)
}

func lowerCounterSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported counter spell",
			"the executable source backend supports only exact counter of one target spell",
		)
	}
	if content, ok := lowerCounterUnlessPaysSpell(ctx); ok {
		return content, nil
	}
	colorGate, hasColorGate := targetColorGateSelection(ctx.content.Conditions)
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		ctx.content.Effects[0].Negated ||
		!ctx.content.Effects[0].Exact ||
		ctx.content.Effects[0].Context != parser.EffectContextController ||
		ctx.content.Effects[0].Amount.Known ||
		len(ctx.content.Conditions) != 0 && !hasColorGate ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return unsupported()
	}
	targetSpec, ok := counterTargetSpec(ctx.content.Targets[0])
	if !ok || len(targetSpec.Predicate.SpellCardTypesAny) != 0 {
		return unsupported()
	}
	instruction := game.Instruction{
		Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)},
	}
	if hasColorGate {
		instruction.Condition = opt.Val(targetColorEffectCondition(
			game.TargetStackObjectReference(0),
			colorGate,
			ctx.content.Conditions[0].Text,
		))
	}
	return game.Mode{
		Targets:  []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{instruction},
	}.Ability(), nil
}

// lowerCounterThenExileInstead lowers the two-effect counter-and-exile body
// "Counter target <filter> spell. If that spell is countered this way, exile it
// instead of putting it into its owner's graveyard." into a single
// CounterObject with ExileInstead set (CR 614 replacement). The parser marks
// the exact exile rider via CounteredSpellExileReplacement; the intrinsic "If
// that spell is countered this way" condition is consumed as part of the
// recognized shape rather than lowered as an independent effect gate.
func lowerCounterThenExileInstead(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	counter := ctx.content.Effects[0]
	exile := ctx.content.Effects[1]
	if counter.Kind != compiler.EffectCounter ||
		!counter.Exact ||
		counter.Negated ||
		counter.Context != parser.EffectContextController ||
		counter.Amount.Known ||
		exile.Kind != compiler.EffectExile ||
		!exile.CounteredSpellExileReplacement {
		return game.AbilityContent{}, false
	}
	if !counterExileRiderConditions(ctx.content.Conditions) {
		return game.AbilityContent{}, false
	}
	target := ctx.content.Targets[0]
	if target.Cardinality.Min != 1 || target.Cardinality.Max != 1 {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := counterTargetSpec(target)
	if !ok || len(targetSpec.Predicate.SpellCardTypesAny) != 0 {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{{
			Primitive: game.CounterObject{
				Object:       game.TargetStackObjectReference(0),
				ExileInstead: true,
			},
		}},
	}.Ability(), true
}

// counterExileRiderConditions reports whether the conditions accompanying a
// counter-and-exile body are exactly the intrinsic "If that spell is countered
// this way" rider (a single plain ConditionIf with no predicate) or none at
// all. Any other condition leaves the body unrecognized so it fails closed.
func counterExileRiderConditions(conditions []compiler.CompiledCondition) bool {
	if len(conditions) == 0 {
		return true
	}
	if len(conditions) != 1 {
		return false
	}
	condition := conditions[0]
	return condition.Kind == compiler.ConditionIf &&
		condition.Predicate == compiler.ConditionPredicateUnsupported &&
		!condition.Negated &&
		!condition.Intervening &&
		!condition.Resolving
}

// lowerChooseNewTargetsSpell lowers the retarget effect "[You may] choose new
// targets for target spell or ability." to a single ChooseNewTargets primitive
// over a stack-object target. The optional "You may" wrapper rides on the
// instruction's Optional flag so the resolving controller decides whether to
// re-choose targets. Any rider (a copy clause, a condition, extra effects)
// leaves the body unrecognized so it fails closed.
func lowerChooseNewTargetsSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported retarget effect",
			"the executable source backend supports only exact retargeting of one target spell or ability",
		)
	}
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return unsupported()
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectChooseNewTargets ||
		!effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextController ||
		effect.Amount.Known ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone {
		return unsupported()
	}
	targetSpec, ok := counterTargetSpec(ctx.content.Targets[0])
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{{
			Primitive: game.ChooseNewTargets{Object: game.TargetStackObjectReference(0)},
			Optional:  effect.Optional || ctx.optional,
		}},
	}.Ability(), nil
}

// lowerChooseCreatureTypeSpell lowers the resolution-time effect "Choose a
// creature type." to a single Choose primitive that publishes the chosen
// subtype under game.SpellChosenTypeChoiceKey. Later effects in the same
// resolution read it through a count selection's SubtypeChoiceResolution ("draw a
// card for each permanent you control of that type", Distant Melody). Any rider
// (target, condition, mode, reference, or extra effect) leaves the body
// unrecognized so it fails closed.
func lowerChooseCreatureTypeSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported choose effect",
			"the executable source backend supports only exact \"Choose a creature type.\"",
		)
	}
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return unsupported()
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectChooseCreatureType ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone {
		return unsupported()
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.Choose{
				Choice: game.ResolutionChoice{
					Kind:          game.ResolutionChoiceSubtype,
					SubtypeOfType: types.Creature,
					Prompt:        "Choose a creature type",
				},
				PublishChoice: game.SpellChosenTypeChoiceKey,
			},
		}},
	}.Ability(), nil
}

// lowerCopyStackObjectSpell lowers "Copy target <activated ability|triggered
// ability|spell, activated ability, or triggered ability|instant or sorcery
// spell|...> [you control][. You may choose new targets for the copy]." to a
// single CopyStackObject primitive over a stack-object target. The optional
// retarget rider (folded by the parser into CopyMayChooseNewTargets) sets
// MayChooseNewTargets so the resolving controller may re-choose the copy's
// targets. Any condition, extra effect, or unrecognized rider leaves the body
// unrecognized so it fails closed.
func lowerCopyStackObjectSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported copy effect",
			"the executable source backend supports only exact copy of one target spell or activated/triggered ability",
		)
	}
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return unsupported()
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectCopyStackObject ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		ctx.optional ||
		effect.Context != parser.EffectContextController ||
		effect.Amount.Known ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone {
		return unsupported()
	}
	object, targets, ok := copyStackObjectReference(ctx)
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{{
			Primitive: game.CopyStackObject{
				Object:              object,
				MayChooseNewTargets: effect.CopyMayChooseNewTargets,
			},
		}},
	}.Ability(), nil
}

// copyStackObjectReference resolves the stack object a copy effect copies along
// with any target specs it needs. The targeted form ("Copy target spell.")
// binds the single stack-object target; the reference form ("copy that spell."
// in a spell-cast trigger) binds the triggering spell through the event and
// needs no targets. It fails closed for any other target/reference shape.
func copyStackObjectReference(ctx contentCtx) (game.ObjectReference, []game.TargetSpec, bool) {
	switch {
	case len(ctx.content.Targets) == 1 && len(ctx.content.References) == 0:
		target := ctx.content.Targets[0]
		if target.Cardinality.Min != 1 || target.Cardinality.Max != 1 {
			return game.ObjectReference{}, nil, false
		}
		targetSpec, ok := counterTargetSpec(target)
		if !ok {
			return game.ObjectReference{}, nil, false
		}
		return game.TargetStackObjectReference(0), []game.TargetSpec{targetSpec}, true
	case len(ctx.content.Targets) == 0 && len(ctx.content.References) == 1:
		object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowEvent: true})
		if !ok || object.Kind() != game.ObjectReferenceEventStackObject {
			return game.ObjectReference{}, nil, false
		}
		return object, nil, true
	default:
		return game.ObjectReference{}, nil, false
	}
}

func lowerSacrificeSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported sacrifice spell",
			"the executable source backend does not yet lower this sacrifice effect",
		)
	}

	if content, ok := lowerSacrificeSourceUnlessPaySpell(ctx); ok {
		return content, nil
	}

	effect := ctx.content.Effects[0]
	if !effect.Exact {
		return unsupported()
	}
	// Source-bound or event-permanent-bound sacrifice of a self-reference: the
	// direct pronoun ("it") or the source object named explicitly ("this
	// creature", the card's own name).
	selfReference := len(ctx.content.References) == 1 &&
		(ctx.content.References[0].Kind == compiler.ReferenceThisObject ||
			ctx.content.References[0].Kind == compiler.ReferenceSelfName ||
			(ctx.content.References[0].Kind == compiler.ReferencePronoun &&
				ctx.content.References[0].Pronoun == compiler.ReferencePronounIt))
	if effect.Context == parser.EffectContextController &&
		len(ctx.content.Targets) == 0 &&
		selfReference &&
		len(ctx.content.Conditions) == 0 &&
		len(ctx.content.Keywords) == 0 &&
		len(ctx.content.Modes) == 0 &&
		!effect.Negated {
		object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
			AllowSource:      true,
			SourceCardObject: true,
			AllowEvent:       true,
		})
		if ok {
			return game.Mode{Sequence: []game.Instruction{{
				Primitive: game.Sacrifice{Object: object},
			}}}.Ability(), nil
		}
	}
	// Strict fail-closed: reject unsupported modifiers and dynamic amounts.
	if len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!effect.Amount.Known ||
		effect.Amount.Value < 1 ||
		effect.Negated {
		return unsupported()
	}

	// Map selector to game.Selection; fail-closed for unknown shapes.
	selection, ok := sacrificeChoiceSelection(effect.Selector)
	if !ok {
		return unsupported()
	}

	amount := game.Fixed(effect.Amount.Value)

	switch {
	case len(ctx.content.Targets) == 1:
		// "Target player/opponent sacrifices <N> <type>."
		target := ctx.content.Targets[0]
		if target.Cardinality.Min != 1 || target.Cardinality.Max != 1 {
			return unsupported()
		}
		targetSpec, ok := playerTargetSpec(target)
		if !ok ||
			effect.Context != parser.EffectContextTarget ||
			!sacrificeChoiceReferences(ctx.content.References) {
			return unsupported()
		}
		return game.Mode{
			Targets: []game.TargetSpec{targetSpec},
			Sequence: []game.Instruction{{
				Primitive: game.SacrificePermanents{
					Player:    game.TargetPlayerReference(0),
					Amount:    amount,
					Selection: selection,
				},
			}},
		}.Ability(), nil

	case len(ctx.content.Targets) == 0:
		// "That player sacrifices <N> <type> of their choice." — the player
		// named by the triggering event (e.g. each opponent's upkeep) chooses.
		if effect.Context == parser.EffectContextReferencedPlayer {
			if !sacrificeReferencedPlayerChoice(ctx.content.References) {
				return unsupported()
			}
			return game.Mode{
				Sequence: []game.Instruction{{
					Primitive: game.SacrificePermanents{
						Player:    game.EventPlayerReference(),
						Amount:    amount,
						Selection: selection,
					},
				}},
			}.Ability(), nil
		}
		// "You sacrifice <N> <type>." or "Each opponent/player sacrifices <N> <type>."
		if !sacrificeChoiceReferences(ctx.content.References) {
			return unsupported()
		}
		if effect.Context == parser.EffectContextController {
			return game.Mode{
				Sequence: []game.Instruction{{
					Primitive: game.SacrificePermanents{
						Player:    game.ControllerReference(),
						Amount:    amount,
						Selection: selection,
					},
				}},
			}.Ability(), nil
		}
		if effect.Context == parser.EffectContextDefendingPlayer {
			return game.Mode{
				Sequence: []game.Instruction{{
					Primitive: game.SacrificePermanents{
						Player:    game.DefendingPlayerReference(),
						Amount:    amount,
						Selection: selection,
					},
				}},
			}.Ability(), nil
		}
		var group game.PlayerGroupReference
		switch effect.Context {
		case parser.EffectContextEachOpponent, parser.EffectContextEachOtherPlayer:
			group = game.OpponentsReference()
		case parser.EffectContextEachPlayer:
			group = game.AllPlayersReference()
		default:
			return unsupported()
		}
		return game.Mode{
			Sequence: []game.Instruction{{
				Primitive: game.SacrificePermanents{
					PlayerGroup: group,
					Amount:      amount,
					Selection:   selection,
				},
			}},
		}.Ability(), nil

	default:
		return unsupported()
	}
}

// lowerSacrificeSourceUnlessPaySpell lowers "sacrifice <this permanent> unless
// you pay {cost}." (Phantasmal Forces, Krosan Cloudscraper, Sunken City, and the
// upkeep "pay or sacrifice" cycle). The controller is offered the fixed mana
// payment as the ability resolves; declining (or being unable to pay) sacrifices
// the source permanent. It is restricted to a single source-bound sacrifice with
// a fixed, non-variable controller payment and no targets, modes, or keywords.
func lowerSacrificeSourceUnlessPaySpell(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Conditions) != 1 ||
		len(ctx.content.References) != 1 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	payment := effect.Payment
	if effect.Kind != compiler.EffectSacrifice ||
		effect.Negated ||
		effect.Context != parser.EffectContextController ||
		payment.Form != parser.EffectPaymentFormUnless ||
		payment.Payer != parser.EffectPaymentPayerController ||
		len(payment.ManaCost) == 0 ||
		manaCostHasVariableSymbol(payment.ManaCost) ||
		payment.GenericManaAmount.DynamicKind != compiler.DynamicAmountNone ||
		ctx.content.Conditions[0].Predicate != compiler.ConditionPredicateControllerDoesNotPay ||
		ctx.content.References[0].Binding != compiler.ReferenceBindingSource {
		return game.AbilityContent{}, false
	}
	const resultKey = game.ResultKey("sacrifice-unless-paid")
	return game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive: game.Pay{Payment: game.ResolutionPayment{
					Prompt:   "Pay " + payment.ManaCost.String() + "?",
					ManaCost: opt.Val(payment.ManaCost),
				}},
				PublishResult: resultKey,
			},
			{
				Primitive: game.Sacrifice{Object: game.SourcePermanentReference()},
				ResultGate: opt.Val(game.InstructionResultGate{
					Key:       resultKey,
					Succeeded: game.TriFalse,
				}),
			},
		},
	}.Ability(), true
}

func sacrificeChoiceReferences(references []compiler.CompiledReference) bool {
	for _, reference := range references {
		if reference.Kind != compiler.ReferencePronoun ||
			reference.Pronoun != compiler.ReferencePronounTheir {
			return false
		}
	}
	return true
}

// sacrificeReferencedPlayerChoice reports whether the references describe a
// "that player sacrifices <type> of their choice" edict: exactly one
// event-player "that player" subject plus zero or more "their"-choice
// possessives. The subject resolves to game.EventPlayerReference, so the player
// named by the triggering event (e.g. each opponent's upkeep) makes the choice.
func sacrificeReferencedPlayerChoice(references []compiler.CompiledReference) bool {
	sawSubject := false
	for _, reference := range references {
		switch {
		case reference.Kind == compiler.ReferenceThatPlayer &&
			reference.Binding == compiler.ReferenceBindingEventPlayer:
			if sawSubject {
				return false
			}
			sawSubject = true
		case reference.Kind == compiler.ReferencePronoun &&
			reference.Pronoun == compiler.ReferencePronounTheir:
		default:
			return false
		}
	}
	return sawSubject
}

// sacrificeChoiceSelection maps the sacrifice effect's compiled selector to a
// runtime Selection. It supports a single permanent card type, a card-type
// union ("creature or planeswalker"), and the nontoken/token qualifier. It
// fails closed for any other selector shape so unrecognized filters stay
// unsupported.
func sacrificeChoiceSelection(selector compiler.CompiledSelector) (game.Selection, bool) {
	var selection game.Selection
	if union := selector.RequiredTypesAny(); len(union) > 1 {
		selection.RequiredTypesAny = union
	} else {
		switch selector.Kind {
		case compiler.SelectorCreature:
			selection.RequiredTypes = []types.Card{types.Creature}
		case compiler.SelectorArtifact:
			selection.RequiredTypes = []types.Card{types.Artifact}
		case compiler.SelectorLand:
			selection.RequiredTypes = []types.Card{types.Land}
		case compiler.SelectorEnchantment:
			selection.RequiredTypes = []types.Card{types.Enchantment}
		case compiler.SelectorPermanent:
			// zero Selection = any permanent
		default:
			return game.Selection{}, false
		}
	}
	switch {
	case selector.NonToken:
		selection.NonToken = true
	case selector.TokenOnly:
		selection.TokenOnly = true
	default:
	}
	// "Sacrifice another creature." sacrifices a permanent other than the
	// effect's own source; the runtime selection drops the source object.
	selection.ExcludeSource = selector.Another || selector.Other
	return selection, true
}

func lowerCounterUnlessPaysSpell(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		ctx.content.Effects[0].Negated ||
		!ctx.content.Effects[0].Exact ||
		len(ctx.content.Conditions) != 1 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!referencesBindTo(ctx.content.References, compiler.ReferenceBindingTarget, 0) {
		return game.AbilityContent{}, false
	}
	payment := ctx.content.Effects[0].Payment
	if payment.Payer != parser.EffectPaymentPayerTargetController ||
		len(payment.ManaCost) == 0 ||
		manaCostHasVariableSymbol(payment.ManaCost) ||
		ctx.content.Conditions[0].Predicate != compiler.ConditionPredicateTargetControllerDoesNotPay {
		return game.AbilityContent{}, false
	}
	target := ctx.content.Targets[0]
	targetSpec, ok := stackSpellTargetSpec(target)
	if !ok {
		return game.AbilityContent{}, false
	}
	const resultKey = game.ResultKey("unless-paid")
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: game.Pay{Payment: game.ResolutionPayment{
					Prompt:   "Pay " + payment.ManaCost.String() + "?",
					Payer:    opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
					ManaCost: opt.Val(payment.ManaCost),
				}},
				PublishResult: resultKey,
			},
			{
				Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)},
				ResultGate: opt.Val(game.InstructionResultGate{
					Key:       resultKey,
					Succeeded: game.TriFalse,
				}),
			},
		},
	}.Ability(), true
}

func playerTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if !target.Exact || !targetCardinalityIsOne(target) {
		return game.TargetSpec{}, false
	}
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: target.Text,
		Allow:      game.TargetAllowPlayer,
	}
	switch target.Selector.Kind {
	case compiler.SelectorPlayer:
	case compiler.SelectorOpponent:
		spec.Predicate = game.TargetPredicate{Player: game.PlayerOpponent}
	default:
		return game.TargetSpec{}, false
	}
	return spec, true
}
