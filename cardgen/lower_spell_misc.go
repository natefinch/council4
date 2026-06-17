package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func lowerFixedLifeSpell(
	ctx contentCtx,
	verb string,
	primitiveFactory func(amount game.Quantity, player game.PlayerReference) game.Primitive,
	groupPrimitiveFactory func(amount game.Quantity, group game.PlayerGroupReference) game.Primitive,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if (effect.Amount.Known && effect.Amount.Value < 1) ||
		effect.Negated ||
		ctx.optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported life spell",
			"the executable source backend supports only exact fixed life changes",
		)
	}
	amount := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	switch {
	case effect.Amount.Known:
		amount = game.Fixed(effect.Amount.Value)
	case effect.Amount.DynamicKind != compiler.DynamicAmountNone:
		dynamic, ok := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		if !ok || effect.Amount.DynamicKind == compiler.DynamicAmountSourcePower ||
			len(ctx.content.References) != 0 {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported life spell",
				"the executable source backend supports only exact supported life changes",
			)
		}
		amount = game.Dynamic(dynamic)
	case len(ctx.content.References) != 0:
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported life spell",
			"the executable source backend supports only exact supported life changes",
		)
	default:
	}
	if !effect.Exact ||
		!effect.Amount.Known && !effect.Amount.VariableX && effect.Amount.DynamicKind == compiler.DynamicAmountNone {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported life spell",
			"the executable source backend supports only exact fixed life changes",
		)
	}
	if len(ctx.content.Targets) == 0 {
		switch effect.Context {
		case parser.EffectContextEachOpponent:
			return game.Mode{
				Sequence: []game.Instruction{{
					Primitive: groupPrimitiveFactory(amount, game.OpponentsReference()),
				}},
			}.Ability(), nil
		case parser.EffectContextEachPlayer:
			return game.Mode{
				Sequence: []game.Instruction{{
					Primitive: groupPrimitiveFactory(amount, game.AllPlayersReference()),
				}},
			}.Ability(), nil
		}
	}
	playerRef := game.ControllerReference()
	var targets []game.TargetSpec
	switch {
	case len(ctx.content.Targets) == 0 &&
		effect.Context == parser.EffectContextController:
	case len(ctx.content.Targets) == 0 &&
		len(ctx.content.References) == 1 &&
		(effect.Context == parser.EffectContextEventPlayer &&
			ctx.content.References[0].Kind == compiler.ReferencePronoun &&
			ctx.content.References[0].Pronoun == compiler.ReferencePronounThey ||
			effect.Context == parser.EffectContextReferencedPlayer &&
				ctx.content.References[0].Kind == compiler.ReferenceThatPlayer &&
				ctx.content.References[0].Binding != compiler.ReferenceBindingTarget):
		playerRef = game.EventPlayerReference()
	case len(ctx.content.Targets) == 1 &&
		effect.Context == parser.EffectContextReferencedObjectController:
		ref, ok := referencedControllerPlayerRef(ctx)
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported life spell",
				"the executable source backend supports only exact fixed life changes",
			)
		}
		playerRef = ref
	case len(ctx.content.Targets) == 1 &&
		(effect.Context == parser.EffectContextTarget || effect.Context == parser.EffectContextPriorSubject):
		targetSpec, ok := playerTargetSpec(ctx.content.Targets[0])
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported life spell",
				"the executable source backend supports only exact fixed life changes",
			)
		}

		targets = []game.TargetSpec{targetSpec}
		playerRef = game.TargetPlayerReference(0)
	default:
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported life spell",
			"the executable source backend supports only exact fixed life changes",
		)
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{{
			Primitive: primitiveFactory(amount, playerRef),
		}},
	}.Ability(), nil
}

func lowerFixedDestroySpell(
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	if group, ok := exactMassDestroyGroup(ctx); ok {
		return game.Mode{
			Sequence: []game.Instruction{
				{
					Primitive: game.Destroy{
						Group: group,
					},
				},
			},
		}.Ability(), nil
	}
	if len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 ||
		ctx.content.Effects[0].Negated ||
		!ctx.content.Effects[0].Exact ||
		ctx.content.Effects[0].Context != parser.EffectContextController {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported destroy spell",
			"the executable source backend supports only exact destruction of one target permanent",
		)
	}
	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported destroy spell",
			"the executable source backend supports only exact destruction of one target permanent",
		)
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: game.Destroy{
					Object: game.TargetPermanentReference(0),
				},
			},
		},
	}.Ability(), nil
}

func lowerFixedExileSpell(
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	if group, ok := exactMassExileGroup(ctx); ok {
		return game.Mode{
			Sequence: []game.Instruction{{
				Primitive: game.Exile{Group: group},
			}},
		}.Ability(), nil
	}
	return lowerFixedPermanentTargetSpell(ctx, "Exile", func(object game.ObjectReference) game.Primitive {
		return game.Exile{Object: object}
	})
}

func exactMassDestroyGroup(ctx contentCtx) (game.GroupReference, bool) {
	return exactMassGroup(ctx)
}

func exactMassExileGroup(ctx contentCtx) (game.GroupReference, bool) {
	return exactMassGroup(ctx)
}

func exactMassGroup(ctx contentCtx) (game.GroupReference, bool) {
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 ||
		ctx.content.Effects[0].Negated ||
		!ctx.content.Effects[0].Exact ||
		!ctx.content.Effects[0].Selector.All {
		return game.GroupReference{}, false
	}
	selection, ok := massGroupSelection(ctx.content.Effects[0].Selector, ctx.content.Keywords)
	if !ok {
		return game.GroupReference{}, false
	}
	if !massGroupKeywordsMatch(ctx.content.Keywords, selection) {
		return game.GroupReference{}, false
	}
	return game.BattlefieldGroup(selection), true
}

func massGroupKeywordsMatch(keywords []compiler.CompiledKeyword, selection game.Selection) bool {
	if selection.Keyword == game.KeywordNone {
		return len(keywords) == 0
	}
	if len(keywords) != 1 || keywords[0].ParameterKind != parser.KeywordParameterNone {
		return false
	}
	keyword, ok := runtimeKeyword(keywords[0].Kind)
	return ok && keyword == selection.Keyword
}

func massGroupSelection(selector compiler.CompiledSelector, keywords []compiler.CompiledKeyword) (game.Selection, bool) {
	selection := game.Selection{
		RequiredTypesAny: append([]types.Card(nil), selector.RequiredTypesAny()...),
		ExcludedTypes:    append([]types.Card(nil), selector.ExcludedTypes()...),
		ColorsAny:        append([]color.Color(nil), selector.ColorsAny()...),
		ExcludedColors:   append([]color.Color(nil), selector.ExcludedColors()...),
		ExcludeSource:    selector.Other,
	}
	if len(selection.RequiredTypesAny) == 0 {
		if requiredType, ok := massGroupRequiredType(selector.Kind); ok {
			selection.RequiredTypes = []types.Card{requiredType}
		} else if selector.Kind != compiler.SelectorPermanent {
			return game.Selection{}, false
		}
	}
	switch selector.Controller {
	case compiler.ControllerAny:
	case compiler.ControllerYou:
		selection.Controller = game.ControllerYou
	case compiler.ControllerOpponent, compiler.ControllerNotYou:
		selection.Controller = game.ControllerOpponent
	default:
		return game.Selection{}, false
	}
	if selector.Tapped {
		selection.Tapped = game.TriTrue
	}
	if selector.MatchManaValue {
		selection.ManaValue = opt.Val(selector.ManaValue)
	}
	if selector.MatchPower {
		selection.Power = opt.Val(selector.Power)
	}
	if selector.MatchToughness {
		selection.Toughness = opt.Val(selector.Toughness)
	}
	if len(keywords) > 0 {
		if len(keywords) != 1 || keywords[0].ParameterKind != parser.KeywordParameterNone {
			return game.Selection{}, false
		}
		keyword, ok := runtimeKeyword(keywords[0].Kind)
		if !ok {
			return game.Selection{}, false
		}
		selection.Keyword = keyword
	}
	return selection, true
}

func massGroupRequiredType(kind compiler.SelectorKind) (types.Card, bool) {
	switch kind {
	case compiler.SelectorArtifact:
		return types.Artifact, true
	case compiler.SelectorCreature:
		return types.Creature, true
	case compiler.SelectorEnchantment:
		return types.Enchantment, true
	case compiler.SelectorLand:
		return types.Land, true
	case compiler.SelectorPlaneswalker:
		return types.Planeswalker, true
	default:
		return "", false
	}
}

func lowerFixedDrawSpell(
	ctx contentCtx,
	_ *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	// Allow a single EventPlayer reference for "They draw N card(s)." bodies;
	// reject all other non-zero-reference forms.
	hasEventPlayerRef := len(ctx.content.References) == 1 &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingEventPlayer
	hasReferencedControllerRef := len(ctx.content.References) == 1 &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingTarget &&
		effect.Context == parser.EffectContextReferencedObjectController
	if (effect.Amount.Known && effect.Amount.Value < 1) ||
		!effect.Amount.Known && !effect.Amount.VariableX && effect.Amount.DynamicKind == compiler.DynamicAmountNone ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		ctx.optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		(len(ctx.content.References) != 0 && !hasEventPlayerRef && !hasReferencedControllerRef) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported draw spell",
			"the executable source backend supports only exact fixed card draw",
		)
	}
	amount := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	if effect.Amount.Known {
		amount = game.Fixed(effect.Amount.Value)
	} else if effect.Amount.DynamicKind != compiler.DynamicAmountNone {
		dynamic, ok := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		if !ok || effect.Amount.DynamicKind == compiler.DynamicAmountSourcePower {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported draw spell",
				"the executable source backend supports only exact supported card draw",
			)
		}
		amount = game.Dynamic(dynamic)
	}
	playerRef := game.ControllerReference()
	var targets []game.TargetSpec
	switch {
	case hasEventPlayerRef && len(ctx.content.Targets) == 0 &&
		(effect.Context == parser.EffectContextEventPlayer || effect.Context == parser.EffectContextReferencedPlayer) &&
		effect.Amount.Known:
		playerRef = game.EventPlayerReference()
	case len(ctx.content.Targets) == 0 &&
		!hasEventPlayerRef &&
		effect.Context == parser.EffectContextController:
	case hasReferencedControllerRef && len(ctx.content.Targets) == 1 && effect.Amount.Known:
		ref, ok := referencedControllerPlayerRef(ctx)
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(ctx, "unsupported draw spell", "the executable source backend supports only exact fixed card draw")
		}
		playerRef = ref
	case len(ctx.content.Targets) == 1 &&
		!hasEventPlayerRef &&
		(effect.Context == parser.EffectContextTarget || effect.Context == parser.EffectContextPriorSubject):
		target, ok := playerTargetSpec(ctx.content.Targets[0])
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported draw spell",
				"the executable source backend supports only exact fixed card draw",
			)
		}
		playerRef = game.TargetPlayerReference(0)
		target.Constraint = "target player"
		targets = []game.TargetSpec{target}
	default:
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported draw spell",
			"the executable source backend supports only exact fixed card draw",
		)
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{
			{
				Primitive: game.Draw{
					Amount: amount,
					Player: playerRef,
				},
			},
		},
	}.Ability(), nil
}

// referencedControllerPlayerRef resolves the recipient player for an "Its
// controller <effect>" body whose subject is the controller of the inherited
// antecedent target in an ordered sequence. The antecedent target's selector
// kind drives the object reference kind: a permanent target yields a permanent
// reference, a spell on the stack yields a stack-object reference (so a
// counterspell's "its controller" resolves the countered spell's controller). It
// returns false (fail closed) for any other shape or antecedent kind. The
// embedded clause-local target index is rebased by the sequence machinery.
func referencedControllerPlayerRef(ctx contentCtx) (game.PlayerReference, bool) {
	if len(ctx.content.Effects) == 0 ||
		ctx.content.Effects[0].Context != parser.EffectContextReferencedObjectController ||
		len(ctx.content.References) != 1 ||
		ctx.content.References[0].Binding != compiler.ReferenceBindingTarget ||
		ctx.content.References[0].Occurrence < 0 ||
		len(ctx.content.Targets) != 1 {
		return game.PlayerReference{}, false
	}
	occ := ctx.content.References[0].Occurrence
	switch ctx.content.Targets[0].Selector.Kind {
	case compiler.SelectorArtifact, compiler.SelectorCreature, compiler.SelectorEnchantment,
		compiler.SelectorLand, compiler.SelectorPermanent, compiler.SelectorPlaneswalker,
		compiler.SelectorBattle:
		return game.ObjectControllerReference(game.TargetPermanentReference(occ)), true
	case compiler.SelectorSpell:
		return game.ObjectControllerReference(game.TargetStackObjectReference(occ)), true
	default:
		return game.PlayerReference{}, false
	}
}
