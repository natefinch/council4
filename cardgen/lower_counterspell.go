package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const delayedDrawAmountChoiceKey = game.ChoiceKey("delayed-draw-amount")

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
	if ctx.optional ||
		len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 1 {
		return game.AbilityContent{}, false
	}
	counterEffect := ctx.content.Effects[0]
	tokenEffect := ctx.content.Effects[1]
	target := ctx.content.Targets[0]
	if counterEffect.Kind != compiler.EffectCounter ||
		counterEffect.Context != parser.EffectContextController ||
		counterEffect.Connection != parser.EffectConnectionNone ||
		!counterEffect.Exact ||
		counterEffect.Optional ||
		counterEffect.Negated ||
		counterEffect.Amount.Known ||
		len(counterEffect.Payment.ManaCost) != 0 ||
		len(counterEffect.Targets) != 1 ||
		!target.Exact ||
		counterEffect.Targets[0].Span != target.Span {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := stackSpellTargetSpec(target)
	if !ok || len(targetSpec.Predicate.SpellCardTypesAny) < 2 {
		return game.AbilityContent{}, false
	}
	if tokenEffect.Kind != compiler.EffectCreate ||
		tokenEffect.Context != parser.EffectContextReferencedObjectController ||
		tokenEffect.Connection != parser.EffectConnectionNone ||
		!tokenEffect.Exact ||
		tokenEffect.Optional ||
		tokenEffect.Negated ||
		tokenEffect.DelayedTiming != 0 ||
		tokenEffect.Duration != compiler.DurationNone ||
		tokenEffect.TokenCopyOfTarget ||
		tokenEffect.TokenName != "" ||
		tokenEffect.Selector.Tapped ||
		tokenEffect.Selector.Attacking ||
		len(tokenEffect.Targets) != 0 ||
		len(tokenEffect.References) != 1 ||
		len(tokenEffect.SubjectReferences) != 1 ||
		!referencesBindTo(tokenEffect.References, compiler.ReferenceBindingTarget, 0) ||
		!referencesBindTo(tokenEffect.SubjectReferences, compiler.ReferenceBindingTarget, 0) {
		return game.AbilityContent{}, false
	}
	def, ok := synthesizeCreatureTokenDef(&tokenEffect, nil)
	if !ok {
		return game.AbilityContent{}, false
	}

	if !tokenEffect.Amount.Known || tokenEffect.Amount.Value != 1 {
		return game.AbilityContent{}, false
	}
	amount := game.Fixed(1)
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)}},
			{Primitive: game.CreateToken{
				Amount:    amount,
				Source:    game.TokenDef(def),
				Recipient: opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
			}},
		},
	}.Ability(), true
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
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		ctx.content.Effects[0].Negated ||
		!ctx.content.Effects[0].Exact ||
		ctx.content.Effects[0].Context != parser.EffectContextController ||
		ctx.content.Effects[0].Amount.Known ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return unsupported()
	}
	targetSpec, ok := counterTargetSpec(ctx.content.Targets[0])
	if !ok || len(targetSpec.Predicate.SpellCardTypesAny) != 0 {
		return unsupported()
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{{
			Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)},
		}},
	}.Ability(), nil
}

func lowerSacrificeSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported sacrifice spell",
			"the executable source backend does not yet lower this sacrifice effect",
		)
	}

	effect := ctx.content.Effects[0]
	if !effect.Exact {
		return unsupported()
	}
	// Source-bound or event-permanent-bound sacrifice of the direct pronoun.
	if effect.Context == parser.EffectContextController &&
		len(ctx.content.Targets) == 0 &&
		len(ctx.content.References) == 1 &&
		ctx.content.References[0].Kind == compiler.ReferencePronoun &&
		ctx.content.References[0].Pronoun == compiler.ReferencePronounIt &&
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

	// Map selector kind to game.Selection; fail-closed for unknown kinds.
	var selection game.Selection
	switch effect.Selector.Kind {
	case compiler.SelectorCreature:
		selection = game.Selection{RequiredTypes: []types.Card{types.Creature}}
	case compiler.SelectorArtifact:
		selection = game.Selection{RequiredTypes: []types.Card{types.Artifact}}
	case compiler.SelectorLand:
		selection = game.Selection{RequiredTypes: []types.Card{types.Land}}
	case compiler.SelectorEnchantment:
		selection = game.Selection{RequiredTypes: []types.Card{types.Enchantment}}
	case compiler.SelectorPermanent:
		// zero Selection = any permanent
	default:
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
		var group game.PlayerGroupReference
		switch effect.Context {
		case parser.EffectContextEachOpponent:
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

func sacrificeChoiceReferences(references []compiler.CompiledReference) bool {
	for _, reference := range references {
		if reference.Kind != compiler.ReferencePronoun ||
			reference.Pronoun != compiler.ReferencePronounTheir {
			return false
		}
	}
	return true
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
