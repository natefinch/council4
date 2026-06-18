package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

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
	if !ok {
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
