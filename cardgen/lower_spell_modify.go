package cardgen

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func lowerGroupDamageSpell(
	_ string,
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Effects) != 1 ||
		effect.Kind != compiler.EffectDealDamage ||
		!effect.Amount.Known ||
		effect.Amount.Value < 1 ||
		effect.Negated ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact fixed group damage amounts",
		)
	}
	damageSource, ok := lowerDamageSourceReference(ctx.content.References)
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact fixed group damage amounts",
		)
	}
	sel := effect.Selector
	var recipient game.DamageRecipient
	switch {
	case sel.Kind == compiler.SelectorOpponent && !sel.Other:
		recipient = game.PlayerGroupDamageRecipient(game.OpponentsReference())
	case sel.Kind == compiler.SelectorPlayer && !sel.Other:
		recipient = game.PlayerGroupDamageRecipient(game.AllPlayersReference())
	case sel.Kind == compiler.SelectorCreature && !sel.Other:
		recipient = game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
		}))
	case sel.Kind == compiler.SelectorCreature && sel.Other:
		recipient = game.GroupDamageRecipient(game.BattlefieldGroupExcluding(
			game.Selection{RequiredTypes: []types.Card{types.Creature}},
			game.SourcePermanentReference(),
		))
	default:
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend does not support this group recipient",
		)
	}
	if !effect.Exact || !exactDamageSourceSyntax(ctx.content.References) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact fixed group damage amounts",
		)
	}
	damage := game.Damage{
		Amount:    game.Fixed(effect.Amount.Value),
		Recipient: recipient,
	}
	if damageSource.Kind() == game.ObjectReferenceEventPermanent {
		damage.DamageSource = opt.Val(damageSource)
	} else if damageSourceIsThisObject(ctx.content.References) {
		damage.DamageSource = opt.Val(game.SourcePermanentReference())
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive: damage,
			},
		},
	}.Ability(), nil
}

func lowerFixedDamageSpell(
	_ string,
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if len(ctx.content.Effects) != 1 ||
		effect.Kind != compiler.EffectDealDamage ||
		!effect.Exact ||
		(effect.Context != parser.EffectContextSource &&
			effect.Context != parser.EffectContextReferencedObject &&
			effect.Context != parser.EffectContextPriorSubject) ||
		(effect.Amount.Known && effect.Amount.Value < 1) ||
		!effect.Amount.Known && !effect.Amount.VariableX && effect.Amount.DynamicKind == compiler.DynamicAmountNone ||
		effect.Negated ||
		len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact supported damage amounts to one target",
		)
	}
	amount := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	var damageSource game.ObjectReference
	var sourceBound bool
	if len(ctx.content.References) > 0 {
		damageSource, sourceBound = lowerDamageSourceReference(ctx.content.References[:1])
	}
	if effect.Amount.Known {
		amount = game.Fixed(effect.Amount.Value)
	} else if effect.Amount.DynamicKind != compiler.DynamicAmountNone {
		amountObject := game.SourcePermanentReference()
		if sourceBound {
			amountObject = damageSource
		}
		dynamic, ok := lowerDynamicAmount(effect.Amount, amountObject)
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported damage spell",
				"the executable source backend supports only exact supported damage amounts to one target",
			)
		}
		amount = game.Dynamic(dynamic)
	}
	target, ok := damageTargetSpec(ctx.content.Targets[0])
	if !ok ||
		!exactDamageSourceSyntax(ctx.content.References) ||
		!exactDamageAmountReferences(effect.Amount, ctx.content.References) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported damage spell",
			"the executable source backend supports only exact supported damage amounts to one target",
		)
	}

	damage := game.Damage{
		Amount:    amount,
		Recipient: game.AnyTargetDamageRecipient(0),
	}
	if sourceBound && damageSource.Kind() == game.ObjectReferenceEventPermanent {
		damage.DamageSource = opt.Val(damageSource)
	} else if damageSourceIsThisObject(ctx.content.References) ||
		effect.Amount.DynamicKind == compiler.DynamicAmountSourcePower {
		damage.DamageSource = opt.Val(game.SourcePermanentReference())
	}
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{
			{
				Primitive: damage,
			},
		},
	}.Ability(), nil
}

// damageSourceIsThisObject reports whether the damage subject is the source
// permanent itself referenced as "this <object>" (ReferenceThisObject bound to
// ReferenceBindingSource). Such damage must carry an explicit
// game.SourcePermanentReference() so the runtime attributes the source
// permanent's keywords (lifelink, deathtouch). The card-name spell form
// (ReferenceSelfName) and the empty default are left unchanged.
func damageSourceIsThisObject(references []compiler.CompiledReference) bool {
	return len(references) > 0 &&
		references[0].Kind == compiler.ReferenceThisObject &&
		references[0].Binding == compiler.ReferenceBindingSource
}

func exactDamageSourceSyntax(references []compiler.CompiledReference) bool {
	if len(references) == 0 {
		return false
	}
	reference := references[0]
	if reference.Kind == compiler.ReferencePronoun && reference.Pronoun == compiler.ReferencePronounIt {
		return reference.Binding == compiler.ReferenceBindingEventPermanent
	}
	if reference.Kind == compiler.ReferenceThisObject {
		return reference.Binding == compiler.ReferenceBindingSource
	}
	return reference.Kind == compiler.ReferenceSelfName
}

func lowerFixedModifyPTSpell(
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := &ctx.content.Effects[0]
	if effect.StaticSubject != compiler.StaticSubjectNone {
		return lowerFixedGroupModifyPTSpell(ctx, effect)
	}
	if len(ctx.content.Targets) == 0 &&
		len(ctx.content.References) == 1 &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingEventPermanent {
		return lowerEventPermanentFixedModifyPT(ctx)
	}
	if len(ctx.content.Targets) == 0 &&
		len(ctx.content.References) == 1 &&
		(ctx.content.References[0].Binding == compiler.ReferenceBindingSource ||
			ctx.content.References[0].Binding == compiler.ReferenceBindingTarget) {
		return lowerReferencedFixedModifyPT(ctx)
	}
	dynamicPT := effect.Amount.DynamicKind != compiler.DynamicAmountNone
	if len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		ctx.content.Targets[0].Selector.Kind != compiler.SelectorCreature ||
		(!dynamicPT && (!effect.PowerDelta.Known || !effect.ToughnessDelta.Known)) ||
		!effect.Exact ||
		effect.Negated ||
		effect.Duration != compiler.DurationUntilEndOfTurn ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!validModifyPTAmount(effect, len(ctx.content.References)) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported power/toughness spell",
			"the executable source backend supports only exact supported target-creature power/toughness changes until end of turn",
		)
	}
	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported power/toughness spell",
			"the executable source backend supports only exact supported target-creature power/toughness changes until end of turn",
		)
	}
	powerDelta := game.Fixed(compiledSignedAmountValue(effect.PowerDelta))
	toughnessDelta := game.Fixed(compiledSignedAmountValue(effect.ToughnessDelta))
	if dynamicPT {
		dynamic, ok := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
		if !ok || effect.Amount.DynamicKind == compiler.DynamicAmountSourcePower {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported power/toughness spell",
				"the executable source backend supports only exact supported target-creature power/toughness changes until end of turn",
			)
		}
		switch effect.Amount.DynamicForm {
		case compiler.DynamicAmountWhereX:
			powerDelta = game.Dynamic(dynamic)
			toughnessDelta = game.Dynamic(dynamic)
		case compiler.DynamicAmountForEach:
			powerDelta = dynamicSignedQuantity(dynamic, effect.PowerDelta)
			toughnessDelta = dynamicSignedQuantity(dynamic, effect.ToughnessDelta)
		default:
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported power/toughness spell",
				"the executable source backend supports only exact supported target-creature power/toughness changes until end of turn",
			)
		}
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: game.ModifyPT{
					Object:         game.TargetPermanentReference(0),
					PowerDelta:     powerDelta,
					ToughnessDelta: toughnessDelta,
					Duration:       game.DurationUntilEndOfTurn,
				},
			},
		},
	}.Ability(), nil
}

// lowerEventPermanentFixedModifyPT lowers an exact fixed until-end-of-turn
// ModifyPT body whose sole non-target subject reference is
// ReferenceBindingEventPermanent. The text must be exactly
// "It gets <power>/<toughness> until end of turn." The object lowers to
// game.EventPermanentReference(), which identifies the permanent named by the
// triggering event.
func lowerEventPermanentFixedModifyPT(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported power/toughness spell",
			"the executable source backend supports only exact fixed until-end-of-turn power/toughness changes to the triggering permanent",
		)
	}
	effect := ctx.content.Effects[0]
	if len(ctx.content.References) != 1 ||
		ctx.content.References[0].Binding != compiler.ReferenceBindingEventPermanent ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!effect.PowerDelta.Known ||
		!effect.ToughnessDelta.Known ||
		!effect.Exact ||
		effect.Negated ||
		effect.Duration != compiler.DurationUntilEndOfTurn ||
		effect.Context != parser.EffectContextReferencedObject {
		return unsupported()
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowEvent: true})
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ModifyPT{
				Object:         object,
				PowerDelta:     game.Fixed(compiledSignedAmountValue(effect.PowerDelta)),
				ToughnessDelta: game.Fixed(compiledSignedAmountValue(effect.ToughnessDelta)),
				Duration:       game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability(), nil
}

// lowerReferencedFixedModifyPT lowers an exact fixed until-end-of-turn ModifyPT
// body whose sole subject reference is the source permanent itself ("This
// creature gets <p>/<t> until end of turn.", EffectContextSource) or a prior
// clause's target referenced by "it" in an ordered sequence ("… It gets <p>/<t>
// until end of turn.", EffectContextReferencedObject). The object lowers to
// game.SourcePermanentReference() or a target reference accordingly.
func lowerReferencedFixedModifyPT(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported power/toughness spell",
			"the executable source backend supports only exact fixed until-end-of-turn power/toughness changes to the source or referenced permanent",
		)
	}
	effect := ctx.content.Effects[0]
	if len(ctx.content.References) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!effect.PowerDelta.Known ||
		!effect.ToughnessDelta.Known ||
		!effect.Exact ||
		effect.Negated ||
		effect.Duration != compiler.DurationUntilEndOfTurn ||
		effect.Amount.DynamicKind != compiler.DynamicAmountNone {
		return unsupported()
	}
	binding := ctx.content.References[0].Binding
	switch {
	case binding == compiler.ReferenceBindingSource && effect.Context == parser.EffectContextSource:
	case binding == compiler.ReferenceBindingTarget && effect.Context == parser.EffectContextReferencedObject:
	default:
		return unsupported()
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		AllowSource: true,
		AllowTarget: true,
	})
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ModifyPT{
				Object:         object,
				PowerDelta:     game.Fixed(compiledSignedAmountValue(effect.PowerDelta)),
				ToughnessDelta: game.Fixed(compiledSignedAmountValue(effect.ToughnessDelta)),
				Duration:       game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability(), nil
}

func lowerFixedGroupModifyPTSpell(
	ctx contentCtx,
	effect *compiler.CompiledEffect,
) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported group power/toughness spell",
			"the executable source backend supports only exact fixed supported group power/toughness changes until end of turn",
		)
	}
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!effect.Exact ||
		effect.Negated ||
		effect.Duration != compiler.DurationUntilEndOfTurn ||
		effect.Amount.DynamicKind != compiler.DynamicAmountNone ||
		!effect.PowerDelta.Known ||
		!effect.ToughnessDelta.Known {
		return unsupported()
	}
	group, ok := resolvingStaticSubjectGroup(effect)
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:          game.LayerPowerToughnessModify,
					Group:          group,
					PowerDelta:     compiledSignedAmountValue(effect.PowerDelta),
					ToughnessDelta: compiledSignedAmountValue(effect.ToughnessDelta),
				}},
				Duration: game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability(), nil
}

func lowerTemporaryKeywordSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported temporary keyword spell",
			"the executable source backend supports only exact non-parameterized keyword grants to one target creature or permanent until end of turn",
		)
	}
	effect := ctx.content.Effects[0]
	referencedObject := len(ctx.content.Targets) == 0 &&
		len(ctx.content.References) == 1 &&
		ctx.content.References[0].Binding == compiler.ReferenceBindingTarget &&
		effect.Context == parser.EffectContextReferencedObject
	targetSubject := len(ctx.content.Targets) == 1 &&
		len(ctx.content.References) == 0 &&
		effect.Context == parser.EffectContextTarget &&
		temporaryKeywordTarget(ctx.content.Targets[0])
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		effect.Kind != compiler.EffectGain ||
		!effect.Exact ||
		(!targetSubject && !referencedObject) ||
		effect.Negated ||
		effect.StaticSubject != compiler.StaticSubjectNone ||
		effect.Duration != compiler.DurationUntilEndOfTurn {
		return unsupported()
	}
	keywords, ok := mixedStaticKeywords(ctx.content.Keywords)
	if !ok {
		return unsupported()
	}
	var object game.ObjectReference
	var target opt.V[game.TargetSpec]
	if targetSubject {
		spec, ok := permanentTargetSpec(ctx.content.Targets[0])
		if !ok {
			return unsupported()
		}
		target = opt.Val(spec)
		object = game.TargetPermanentReference(0)
	} else {
		object, ok = lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowTarget: true})
		if !ok {
			return unsupported()
		}
	}
	mode := game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				Object: opt.Val(object),
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:       game.LayerAbility,
					AddKeywords: keywords,
				}},
				Duration: game.DurationUntilEndOfTurn,
			},
		}},
	}
	if target.Exists {
		mode.Targets = []game.TargetSpec{target.Val}
	}
	return mode.Ability(), nil
}

func lowerTemporaryPTKeywordSpell(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		ctx.content.Effects[0].Kind != compiler.EffectModifyPT ||
		ctx.content.Effects[1].Kind != compiler.EffectGain ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!temporaryKeywordTarget(ctx.content.Targets[0]) {
		return game.AbilityContent{}, false
	}
	modifyEffect := ctx.content.Effects[0]
	keywordEffect := ctx.content.Effects[1]
	if modifyEffect.Span != keywordEffect.Span ||
		!modifyEffect.Exact ||
		!keywordEffect.Exact ||
		modifyEffect.Negated ||
		keywordEffect.Negated ||
		modifyEffect.StaticSubject != compiler.StaticSubjectNone ||
		keywordEffect.StaticSubject != compiler.StaticSubjectNone ||
		modifyEffect.Duration != compiler.DurationUntilEndOfTurn ||
		keywordEffect.Duration != compiler.DurationUntilEndOfTurn ||
		modifyEffect.Amount.DynamicKind != compiler.DynamicAmountNone ||
		!modifyEffect.PowerDelta.Known ||
		!modifyEffect.ToughnessDelta.Known {
		return game.AbilityContent{}, false
	}
	keywords, ok := mixedStaticKeywords(ctx.content.Keywords)
	if !ok {
		return game.AbilityContent{}, false
	}
	target, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				Object: opt.Val(game.TargetPermanentReference(0)),
				ContinuousEffects: []game.ContinuousEffect{
					{
						Layer:          game.LayerPowerToughnessModify,
						PowerDelta:     compiledSignedAmountValue(modifyEffect.PowerDelta),
						ToughnessDelta: compiledSignedAmountValue(modifyEffect.ToughnessDelta),
					},
					{
						Layer:       game.LayerAbility,
						AddKeywords: keywords,
					},
				},
				Duration: game.DurationUntilEndOfTurn,
			},
		}},
	}.Ability(), true
}

func temporaryKeywordTarget(target compiler.CompiledTarget) bool {
	return target.Selector.Kind == compiler.SelectorCreature ||
		target.Selector.Kind == compiler.SelectorPermanent
}

func lowerFixedBounceSpell(
	ctx contentCtx,
) (game.AbilityContent, *shared.Diagnostic) {
	if len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		ctx.content.Effects[0].Negated ||
		ctx.content.Effects[0].Optional ||
		!ctx.content.Effects[0].Exact ||
		ctx.optional ||
		ctx.content.Effects[0].Context != parser.EffectContextController ||
		ctx.content.Effects[0].ToZone != zone.Hand ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!referencesBindTo(ctx.content.References, compiler.ReferenceBindingTarget, 0) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported return spell",
			"the executable source backend supports only exact return of one target permanent to its owner's hand",
		)
	}
	target := ctx.content.Targets[0]
	targetSpec, ok := permanentTargetSpec(target)
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported return spell",
			"the executable source backend supports only exact return of one target permanent to its owner's hand",
		)
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowTarget: true})
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported return spell",
			"the executable source backend supports only exact return of one target permanent to its owner's hand",
		)
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: game.Bounce{
					Object: object,
				},
			},
		},
	}.Ability(), nil
}

func lowerFixedPermanentTargetSpell(
	ctx contentCtx,
	verb string,
	primitiveFactory func(object game.ObjectReference) game.Primitive,
) (game.AbilityContent, *shared.Diagnostic) {
	if len(ctx.content.Targets) != 1 ||
		ctx.content.Targets[0].Cardinality.Min != 1 ||
		ctx.content.Targets[0].Cardinality.Max != 1 ||
		ctx.content.Effects[0].Negated ||
		ctx.content.Effects[0].Optional ||
		!ctx.content.Effects[0].Exact ||
		ctx.optional ||
		ctx.content.Effects[0].Context != parser.EffectContextController ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+strings.ToLower(verb)+" spell",
			"the executable source backend supports only exact "+strings.ToLower(verb)+" of one target permanent",
		)
	}
	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+strings.ToLower(verb)+" spell",
			"the executable source backend supports only exact "+strings.ToLower(verb)+" of one target permanent",
		)
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{
				Primitive: primitiveFactory(game.TargetPermanentReference(0)),
			},
		},
	}.Ability(), nil
}

func lowerFixedCardCountPlayerSpell(
	ctx contentCtx,
	_ *parser.Ability,
	controllerVerb string,
	targetVerb string,
	allowDynamic bool,
	primitiveFactory func(amount game.Quantity, player game.PlayerReference) game.Primitive,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	// Allow a single EventPlayer reference for "They {verb} N card(s)." bodies;
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
		effect.Selector.Kind != compiler.SelectorCard ||
		ctx.optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		(len(ctx.content.References) != 0 && !hasEventPlayerRef && !hasReferencedControllerRef) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+controllerVerb+" spell",
			"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
		)
	}
	amount, ok := cardCountQuantity(effect.Amount, allowDynamic)
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+controllerVerb+" spell",
			"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
		)
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
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported "+controllerVerb+" spell",
				"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
			)
		}
		playerRef = ref
	case len(ctx.content.Targets) == 1 &&
		!hasEventPlayerRef &&
		(effect.Context == parser.EffectContextTarget || effect.Context == parser.EffectContextPriorSubject):
		targetSpec, ok := playerTargetSpec(ctx.content.Targets[0])
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported "+controllerVerb+" spell",
				"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
			)
		}
		playerRef = game.TargetPlayerReference(0)
		targets = []game.TargetSpec{targetSpec}
	default:
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+controllerVerb+" spell",
			"the executable source backend supports only exact fixed "+controllerVerb+" by one player",
		)
	}
	return game.Mode{
		Targets: targets,
		Sequence: []game.Instruction{
			{
				Primitive: primitiveFactory(amount, playerRef),
			},
		},
	}.Ability(), nil
}

func lowerFixedControllerSpell(
	ctx contentCtx,
	_ *parser.Ability,
	verb string,
	allowDynamic bool,
	primitiveFactory func(amount game.Quantity, player game.PlayerReference) game.Primitive,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if (effect.Amount.Known && effect.Amount.Value < 1) ||
		!effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextController ||
		ctx.content.Unconsumed() ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+verb+" spell",
			"the executable source backend supports only exact fixed controller "+verb,
		)
	}
	amount, ok := controllerActionQuantity(effect.Amount, allowDynamic)
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported "+verb+" spell",
			"the executable source backend supports only exact fixed controller "+verb,
		)
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{
				Primitive: primitiveFactory(amount, game.ControllerReference()),
			},
		},
	}.Ability(), nil
}

func cardCountQuantity(amount compiler.CompiledAmount, allowDynamic bool) (game.Quantity, bool) {
	if amount.Known {
		return game.Fixed(amount.Value), amount.Value > 0
	}
	if !allowDynamic {
		return game.Quantity{}, false
	}
	if amount.DynamicKind == compiler.DynamicAmountNone {
		return game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}), amount.VariableX
	}
	dynamic, ok := lowerDynamicAmount(amount, game.SourcePermanentReference())
	if !ok || amount.DynamicKind == compiler.DynamicAmountSourcePower {
		return game.Quantity{}, false
	}
	return game.Dynamic(dynamic), true
}

func controllerActionQuantity(amount compiler.CompiledAmount, allowDynamic bool) (game.Quantity, bool) {
	if amount.Known {
		return game.Fixed(amount.Value), amount.Value > 0
	}
	if !allowDynamic {
		return game.Quantity{}, false
	}
	if amount.DynamicKind == compiler.DynamicAmountNone {
		return game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}), amount.VariableX
	}
	dynamic, ok := lowerDynamicAmount(amount, game.SourcePermanentReference())
	if !ok || amount.DynamicKind == compiler.DynamicAmountSourcePower {
		return game.Quantity{}, false
	}
	return game.Dynamic(dynamic), true
}
