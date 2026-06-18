package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// contentCtx is the internal lowering context for ability body content.
// It holds the normalized body text (for exact-pattern matching), the source
// span (for diagnostic attribution), the optional flag, and the compiled
// semantic content. It is NOT an compiler.CompiledAbility and carries no shell
// semantics.
type contentCtx struct {
	text     string
	span     shared.Span
	optional bool
	content  compiler.AbilityContent
	// sequenceClause is true when this context is one sub-clause of a
	// multi-effect ordered sequence (built by contextForEffect). It gates
	// EventPermanent "it"/"that creature" counter placement: in a sequence the
	// compiler binds a pronoun whose antecedent is a prior instruction's product
	// (e.g. a created token) to the triggering event permanent, so accepting it
	// would place counters on the wrong object. Standalone effects keep the
	// EventPermanent binding, which always denotes the triggering permanent.
	sequenceClause bool
	// triggerCardCountEvent is the draw or discard event kind of the enclosing
	// "one or more" trigger, or game.EventUnknown outside such a trigger. It
	// gates DynamicAmountEventCardCount amounts ("for each card discarded this
	// way") so they resolve only against a matching triggering event.
	triggerCardCountEvent game.EventKind
}

// contentDiagnostic creates a content-level diagnostic attributed to ctx.span.
func contentDiagnostic(ctx contentCtx, summary, detail string) *shared.Diagnostic {
	return &shared.Diagnostic{
		Severity: shared.SeverityWarning,
		Summary:  summary,
		Detail:   detail,
		Span:     ctx.span,
	}
}

// lowerAbilityContent is the single entry point for lowering oracle semantic
// content (targets, conditions, effects, keywords, references) into a
// game.AbilityContent value. All ability shells (spell, activated body,
// triggered body, loyalty body, chapter body, and modal option) route their
// body content through this function. Shell lowerers do not create fake
// AbilitySpell wrappers; they build the adjusted content and body syntax
// directly and call this function.
func lowerAbilityContent(
	cardName string,
	content compiler.AbilityContent,
	optional bool,
	bodySyntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	ctx := contentCtx{
		text:     bodySyntax.Text,
		span:     bodySyntax.Span,
		optional: optional,
		content:  content,
	}
	return lowerContent(cardName, ctx, bodySyntax)
}

// lowerSequenceClauseContent lowers one sub-clause of a multi-effect ordered
// sequence, marking the context as a sequence clause. The marker gates pronoun
// bindings (such as EventPermanent "it"/"that creature" counter placement) that
// are only trustworthy for a standalone effect: within a sequence the compiler
// may bind a pronoun whose antecedent is a prior instruction's product to the
// triggering event permanent.
func lowerSequenceClauseContent(
	cardName string,
	content compiler.AbilityContent,
	optional bool,
	bodySyntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	ctx := contentCtx{
		text:           bodySyntax.Text,
		span:           bodySyntax.Span,
		optional:       optional,
		content:        content,
		sequenceClause: true,
	}
	return lowerContent(cardName, ctx, bodySyntax)
}

// lowerTriggerBodyContent lowers a triggered ability body while recording the
// triggering draw or discard event kind, enabling DynamicAmountEventCardCount
// amounts ("for each card discarded this way") that read the triggering event's
// card count. triggerEvent must be the trigger's draw/discard/cycle event kind.
func lowerTriggerBodyContent(
	cardName string,
	content compiler.AbilityContent,
	optional bool,
	bodySyntax *parser.Ability,
	triggerEvent game.EventKind,
) (game.AbilityContent, *shared.Diagnostic) {
	ctx := contentCtx{
		text:                  bodySyntax.Text,
		span:                  bodySyntax.Span,
		optional:              optional,
		content:               content,
		triggerCardCountEvent: triggerEvent,
	}
	return lowerContent(cardName, ctx, bodySyntax)
}

func lowerContent(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	if hasOptionalResolvingEffect(ctx.content.Effects) {
		// Resolving optionality ("you may X. If you do, Y") is lowered only
		// through the ordered effect-sequence path, which wires the optional
		// instruction and its result gate. Any other shape (modal, search,
		// manifest, or a single optional effect) remains unsupported and fails
		// closed.
		if len(ctx.content.Modes) == 0 &&
			len(ctx.content.Effects) > 1 &&
			ctx.content.Effects[0].Kind != compiler.EffectSearch &&
			!typedManifestDreadSequence(ctx.content) {
			if content, diagnostic := lowerOrderedEffectSequence(cardName, ctx, syntax); diagnostic == nil {
				return content, nil
			}
		}
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported optional effect",
			"the executable source backend does not yet lower optional resolving effects",
		)
	}
	if len(ctx.content.Modes) > 0 {
		return lowerModalContent(cardName, ctx, syntax)
	}
	if content, ok := lowerEventCardEffect(ctx); ok {
		return content, nil
	}
	if typedManifestDreadSequence(ctx.content) {
		return manifestDreadAbility(), nil
	}
	if len(ctx.content.Effects) > 0 && ctx.content.Effects[0].Kind == compiler.EffectSearch {
		return lowerSearchSpell(ctx)
	}
	if len(ctx.content.Effects) > 1 {
		if ctx.content.Effects[0].Kind == compiler.EffectGainControl ||
			(ctx.content.Effects[0].Kind == compiler.EffectUntap &&
				len(ctx.content.Effects) >= 2 &&
				ctx.content.Effects[1].Kind == compiler.EffectGainControl) {
			return lowerControlSpellSequence(cardName, ctx, syntax)
		}
		return lowerOrderedEffectSequence(cardName, ctx, syntax)
	}
	if len(ctx.content.Effects) == 1 {
		if ctx.content.Effects[0].RequiresOrderedLowering {
			return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — single effect requires ordered lowering")
		}
		if ctx.content.Effects[0].Kind == compiler.EffectAddMana {
			return lowerAddManaContent(ctx)
		}
		return lowerSingleEffectSpell(cardName, ctx, syntax)
	}
	return game.AbilityContent{}, contentDiagnostic(
		ctx,
		"unsupported ability content",
		"the executable source backend does not yet lower this ability content",
	)
}

func hasOptionalResolvingEffect(effects []compiler.CompiledEffect) bool {
	for i := range effects {
		if effects[i].Optional {
			return true
		}
	}
	return false
}

func lowerSearchSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func(detail string) (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported search effect",
			detail,
		)
	}
	// Search is one runtime primitive, but each reference still binds to the
	// prior semantic search/reveal instruction that produced the found card.
	for _, ref := range ctx.content.References {
		if ref.Binding != compiler.ReferenceBindingPriorInstructionResult {
			return unsupported("unexpected non-result reference in search effect")
		}
	}
	consumed := ctx
	consumed.content.References = nil
	if ctx.optional ||
		consumed.content.Unconsumed() ||
		!exactSearchEffectSequence(ctx.content.Effects) {
		return unsupported("the executable source backend supports only exact unconditional library-search sequences")
	}
	search := ctx.content.Effects[0]
	if !search.Amount.Known || search.Amount.Value != 1 {
		return unsupported("the executable source backend supports only searches for exactly one card")
	}
	for i := range ctx.content.Effects {
		if ctx.content.Effects[i].Span != search.Span ||
			ctx.content.Effects[i].DelayedTiming != 0 ||
			ctx.content.Effects[i].Duration != compiler.DurationNone ||
			ctx.content.Effects[i].Negated {
			return unsupported("the executable source backend supports only exact same-sentence library-search sequences")
		}
	}
	if search.UnsupportedDetail != "" {
		return unsupported(search.UnsupportedDetail)
	}
	if search.Context != parser.EffectContextController ||
		ctx.content.Effects[len(ctx.content.Effects)-1].Connection != parser.EffectConnectionThen {
		return unsupported("the executable source backend supports only searches of your library ending with \"then shuffle\"")
	}
	spec, ok := searchSpecForSelector(search.Selector)
	if !ok {
		return unsupported("unsupported library-search filter")
	}
	spec.SourceZone = zone.Library

	spec.Reveal = len(ctx.content.Effects) == 4
	putIndex := 1
	if spec.Reveal {
		putIndex = 2
	}
	put := ctx.content.Effects[putIndex]
	if put.ToZone != zone.Hand && put.ToZone != zone.Battlefield {
		return unsupported("the executable source backend supports only exact hand or battlefield search destinations")
	}
	spec.Destination = put.ToZone
	spec.EntersTapped = put.EntersTapped

	return game.Mode{Sequence: []game.Instruction{{Primitive: game.Search{
		Player: game.ControllerReference(),
		Spec:   spec,
		Amount: game.Fixed(1),
	}}}}.Ability(), nil
}

func exactSearchEffectSequence(effects []compiler.CompiledEffect) bool {
	if len(effects) == 3 {
		return effects[0].Kind == compiler.EffectSearch &&
			effects[1].Kind == compiler.EffectPut &&
			effects[2].Kind == compiler.EffectShuffle
	}
	return len(effects) == 4 &&
		effects[0].Kind == compiler.EffectSearch &&
		effects[1].Kind == compiler.EffectReveal &&
		effects[2].Kind == compiler.EffectPut &&
		effects[3].Kind == compiler.EffectShuffle
}

func searchSpecForSelector(selector compiler.CompiledSelector) (game.SearchSpec, bool) {
	var spec game.SearchSpec
	if selector.Controller != compiler.ControllerAny ||
		selector.All ||
		selector.Another ||
		selector.Other ||
		selector.Attacking ||
		selector.Blocking ||
		selector.Tapped ||
		selector.Untapped ||
		selector.Keyword != parser.KeywordUnknown ||
		selector.Zone != zone.None ||
		selector.MatchManaValue ||
		selector.MatchPower ||
		selector.MatchToughness ||
		len(selector.RequiredTypesAny()) != 0 ||
		len(selector.ExcludedTypes()) != 0 ||
		len(selector.ColorsAny()) != 0 ||
		len(selector.ExcludedColors()) != 0 {
		return game.SearchSpec{}, false
	}
	switch selector.Kind {
	case compiler.SelectorCard:
	case compiler.SelectorLand:
		spec.CardType = opt.Val(types.Land)
	case compiler.SelectorCreature:
		spec.CardType = opt.Val(types.Creature)
	case compiler.SelectorArtifact:
		spec.CardType = opt.Val(types.Artifact)
	case compiler.SelectorEnchantment:
		spec.CardType = opt.Val(types.Enchantment)
	default:
		return game.SearchSpec{}, false
	}
	supertypes := selector.Supertypes()
	if len(supertypes) > 1 || len(supertypes) == 1 && supertypes[0] != types.Basic {
		return game.SearchSpec{}, false
	}
	if len(supertypes) == 1 {
		spec.Supertype = opt.Val(types.Basic)
	}
	spec.SubtypesAny = slices.Clone(selector.SubtypesAny())
	return spec, true
}

func lowerSingleEffectSpell(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	if len(ctx.content.Effects) == 1 && ctx.content.Effects[0].DelayedTiming != 0 {
		return lowerDelayedSingleEffectSpell(cardName, ctx, syntax)
	}
	return lowerImmediateSingleEffectSpell(cardName, ctx, syntax)
}

func lowerDelayedSingleEffectSpell(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	ctx.content.Effects[0].DelayedTiming = 0

	var content game.AbilityContent
	if primitive, ok := lowerDelayedSelfPrimitive(ctx); ok {
		content = game.Mode{Sequence: []game.Instruction{{Primitive: primitive}}}.Ability()
	} else {
		var diagnostic *shared.Diagnostic
		content, diagnostic = lowerImmediateSingleEffectSpell(cardName, ctx, syntax)
		if diagnostic != nil {
			return game.AbilityContent{}, unsupportedDelayedEffectDiagnostic(ctx)
		}
	}
	if len(content.SharedTargets) != 0 ||
		content.IsModal() ||
		len(content.Modes) != 1 ||
		len(content.Modes[0].Targets) != 0 ||
		len(content.Modes[0].Sequence) == 0 {
		return game.AbilityContent{}, unsupportedDelayedEffectDiagnostic(ctx)
	}
	return game.Mode{Sequence: []game.Instruction{{Primitive: game.CreateDelayedTrigger{
		Trigger: game.DelayedTriggerDef{
			Timing:  effect.DelayedTiming,
			Content: content,
		},
	}}}}.Ability(), nil
}

func lowerDelayedSelfPrimitive(ctx contentCtx) (game.Primitive, bool) {
	if ctx.content.Effects[0].Negated {
		return nil, false
	}
	if !referencesBindTo(ctx.content.References, compiler.ReferenceBindingSource, 0) {
		return nil, false
	}
	consumed := ctx
	consumed.content.References = nil
	if consumed.content.Unconsumed() {
		return nil, false
	}
	sourcePermanent, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		AllowSource:      true,
		SourceCardObject: true,
	})
	if !ok {
		return nil, false
	}
	effect := ctx.content.Effects[0]
	switch effect.Kind {
	case compiler.EffectExile:
		return game.Exile{Object: sourcePermanent}, true
	case compiler.EffectSacrifice:
		return game.Sacrifice{Object: sourcePermanent}, true
	case compiler.EffectReturn:
		if effect.ToZone != zone.Hand {
			return nil, false
		}
		sourceCard, ok := lowerCardReference(ctx.content.References[0], referenceLoweringContext{AllowSource: true})
		if !ok {
			return nil, false
		}
		return game.MoveCard{
			Card:        sourceCard,
			FromZone:    zone.Graveyard,
			Destination: zone.Hand,
		}, true
	default:
		return nil, false
	}
}

func unsupportedDelayedEffectDiagnostic(ctx contentCtx) *shared.Diagnostic {
	return contentDiagnostic(
		ctx,
		"unsupported delayed effect",
		"the executable source backend supports only exact non-target delayed one-shot effects",
	)
}

// lowerReferencedPronounEffect lowers a no-target single effect whose object is
// an "it"/"its" pronoun bound either to the triggering permanent
// (ReferenceBindingEventPermanent) or to a prior clause's target in an ordered
// sequence (ReferenceBindingTarget). It covers destroy, exile, tap, untap,
// sacrifice, and return-to-hand. The object lowers to the event-permanent or a
// target reference accordingly.
func lowerReferencedPronounEffect(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) == 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		ctx.content.Effects[0].Negated ||
		!ctx.content.Effects[0].Exact ||
		ctx.content.Effects[0].Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	hasDirectPronoun := false
	for _, ref := range ctx.content.References {
		if (ref.Binding != compiler.ReferenceBindingEventPermanent &&
			ref.Binding != compiler.ReferenceBindingTarget) ||
			ref.Kind != compiler.ReferencePronoun ||
			(ref.Pronoun != compiler.ReferencePronounIt && ref.Pronoun != compiler.ReferencePronounIts) {
			return game.AbilityContent{}, false
		}
		hasDirectPronoun = hasDirectPronoun || ref.Pronoun == compiler.ReferencePronounIt
	}
	if !hasDirectPronoun {
		return game.AbilityContent{}, false
	}
	consumed := ctx
	consumed.content.References = nil
	if consumed.content.Unconsumed() {
		return game.AbilityContent{}, false
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		AllowEvent:  true,
		AllowTarget: true,
	})
	if !ok {
		return game.AbilityContent{}, false
	}
	var primitive game.Primitive
	effect := ctx.content.Effects[0]
	switch effect.Kind {
	case compiler.EffectDestroy:
		primitive = game.Destroy{Object: object}
	case compiler.EffectExile:
		primitive = game.Exile{Object: object}
	case compiler.EffectTap:
		primitive = game.Tap{Object: object}
	case compiler.EffectUntap:
		primitive = game.Untap{Object: object}
	case compiler.EffectSacrifice:
		primitive = game.Sacrifice{Object: object}
	case compiler.EffectReturn:
		if effect.ToZone != zone.Hand {
			return game.AbilityContent{}, false
		}
		primitive = game.Bounce{Object: object}
	default:
		return game.AbilityContent{}, false
	}
	return game.Mode{Sequence: []game.Instruction{{Primitive: primitive}}}.Ability(), true
}

func lowerImmediateSingleEffectSpell(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	ctx.text = textWithoutDelimited(ctx.text, ctx.span, syntax.Reminders)
	syntax.Tokens = slices.DeleteFunc(
		append([]shared.Token(nil), syntax.Tokens...),
		func(token shared.Token) bool {
			return spanCoveredByDelimited(token.Span, syntax.Reminders)
		},
	)
	// Route no-target EventPermanent pronoun bodies through the shared path
	// before individual effect dispatch so all compatible trigger shells
	// benefit from the same lowering.
	if content, ok := lowerReferencedPronounEffect(ctx); ok {
		return content, nil
	}
	switch ctx.content.Effects[0].Kind {
	case compiler.EffectDealDamage:
		if ctx.content.Effects[0].Divided {
			return lowerDividedDamageSpell(ctx)
		}
		if content, ok := lowerInheritedPowerDamageSpell(ctx); ok {
			return content, nil
		}
		if len(ctx.content.Targets) == 0 {
			return lowerGroupDamageSpell(cardName, ctx)
		}
		return lowerFixedDamageSpell(cardName, ctx)
	case compiler.EffectDraw:
		return lowerFixedDrawSpell(ctx, syntax)
	case compiler.EffectDestroy:
		return lowerFixedDestroySpell(ctx)
	case compiler.EffectGain:
		if len(ctx.content.Keywords) != 0 &&
			ctx.content.Effects[0].Duration == compiler.DurationUntilEndOfTurn {
			return lowerTemporaryKeywordSpell(ctx)
		}
		if !ctx.content.Effects[0].LifeObject {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported keyword or ability grant",
				"the executable source backend does not yet lower spells that grant a keyword or quoted ability",
			)
		}
		return lowerFixedLifeSpell(ctx, "gain", func(amount game.Quantity, player game.PlayerReference) game.Primitive {
			return game.GainLife{Amount: amount, Player: player}
		}, func(amount game.Quantity, group game.PlayerGroupReference) game.Primitive {
			return game.GainLife{Amount: amount, PlayerGroup: group}
		})
	case compiler.EffectGainControl:
		return lowerSingleControlSpell(ctx)
	case compiler.EffectLose:
		if !ctx.content.Effects[0].LifeObject {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported keyword or ability loss",
				"the executable source backend does not yet lower spells that remove a keyword or ability",
			)
		}
		return lowerFixedLifeSpell(ctx, "lose", func(amount game.Quantity, player game.PlayerReference) game.Primitive {
			return game.LoseLife{Amount: amount, Player: player}
		}, func(amount game.Quantity, group game.PlayerGroupReference) game.Primitive {
			return game.LoseLife{Amount: amount, PlayerGroup: group}
		})
	case compiler.EffectScry:
		return lowerFixedControllerSpell(ctx, syntax, "scry", false, func(amount game.Quantity, player game.PlayerReference) game.Primitive {
			return game.Scry{Amount: amount, Player: player}
		})
	case compiler.EffectSurveil:
		return lowerFixedControllerSpell(ctx, syntax, "surveil", false, func(amount game.Quantity, player game.PlayerReference) game.Primitive {
			return game.Surveil{Amount: amount, Player: player}
		})
	case compiler.EffectInvestigate:
		return lowerInvestigateSpell(ctx, syntax)
	case compiler.EffectProliferate:
		return lowerExactPrimitiveSpell(ctx, syntax, "proliferate", func(amount game.Quantity) game.Primitive {
			return game.Proliferate{Amount: amount}
		})
	case compiler.EffectExplore:
		return lowerExploreSpell(ctx)
	case compiler.EffectManifest, compiler.EffectManifestDread:
		return lowerManifestSpell(ctx)
	case compiler.EffectRegenerate:
		return lowerFixedPermanentTargetSpell(ctx, "Regenerate", func(object game.ObjectReference) game.Primitive {
			return game.Regenerate{Object: object}
		})
	case compiler.EffectFight:
		return lowerFightSpell(ctx)
	case compiler.EffectDiscard:
		return lowerFixedCardCountPlayerSpell(
			ctx, syntax, "discard", "discards", false, func(amount game.Quantity, player game.PlayerReference) game.Primitive {
				return game.Discard{Amount: amount, Player: player}
			},
		)
	case compiler.EffectMill:
		return lowerFixedCardCountPlayerSpell(
			ctx, syntax, "mill", "mills", true, func(amount game.Quantity, player game.PlayerReference) game.Primitive {
				return game.Mill{Amount: amount, Player: player}
			},
		)
	case compiler.EffectTap:
		return lowerFixedPermanentTargetSpell(ctx, "Tap", func(object game.ObjectReference) game.Primitive {
			return game.Tap{Object: object}
		})
	case compiler.EffectUntap:
		return lowerFixedPermanentTargetSpell(ctx, "Untap", func(object game.ObjectReference) game.Primitive {
			return game.Untap{Object: object}
		})
	case compiler.EffectExile:
		return lowerFixedExileSpell(ctx)
	case compiler.EffectReturn:
		if content, ok := lowerSelfCardGraveyardReturn(ctx); ok {
			return content, nil
		}
		if content, ok := lowerTargetedGraveyardReturn(ctx); ok {
			return content, nil
		}
		if content, ok := lowerMultiTargetBounceSpell(ctx); ok {
			return content, nil
		}
		return lowerFixedBounceSpell(ctx)
	case compiler.EffectPut:
		if content, ok := lowerTargetedGraveyardReturn(ctx); ok {
			return content, nil
		}
		if ctx.content.Effects[0].ToZone == zone.Library {
			return game.AbilityContent{}, unsupportedLibraryPlacementDiagnostic(ctx)
		}
		return lowerCounterPlacementSpell(ctx)
	case compiler.EffectModifyPT:
		return lowerFixedModifyPTSpell(ctx, syntax)
	case compiler.EffectCounter:
		return lowerCounterSpell(ctx)
	case compiler.EffectSacrifice:
		return lowerSacrificeSpell(ctx)
	case compiler.EffectCreate:
		return lowerCreateTokenSpell(ctx)
	default:
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported ability content",
			"the executable source backend does not yet lower this ability content",
		)
	}
}
