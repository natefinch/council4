package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// contentCtx is the internal lowering context for ability body content.
// It holds the normalized body text (for exact-pattern matching), the source
// span (for diagnostic attribution), the optional flag, the compiled semantic
// content, and the typed enclosing ability kind.
type contentCtx struct {
	text          string
	span          shared.Span
	optional      bool
	content       compiler.AbilityContent
	enclosingKind compiler.AbilityKind
	// sequenceClause is true when this context is one sub-clause of a
	// multi-effect ordered sequence (built by contextForEffect). It gates
	// EventPermanent "it"/"that creature" counter placement: in a sequence the
	// compiler binds a pronoun whose antecedent is a prior instruction's product
	// (e.g. a created token) to the triggering event permanent, so accepting it
	// would place counters on the wrong object. Standalone effects keep the
	// EventPermanent binding, which always denotes the triggering permanent.
	sequenceClause bool
	// allowEventPronoun re-permits an EventPermanent "it"/"that creature"
	// reference inside a sequence clause that is a mutually-exclusive branch
	// (an "Otherwise," else branch). Such a branch never resolves alongside its
	// sibling, so its pronoun cannot denote the sibling's product and safely
	// binds the triggering permanent.
	allowEventPronoun bool
	// triggerCardCountEvent is the draw or discard event kind of the enclosing
	// "one or more" trigger, or game.EventUnknown outside such a trigger. It
	// gates DynamicAmountEventCardCount amounts ("for each card discarded this
	// way") so they resolve only against a matching triggering event.
	triggerCardCountEvent game.EventKind
	// triggerEvent is the enclosing trigger's event kind, or game.EventUnknown
	// outside a triggered ability. It lets typed event-player references lower
	// only where the resolving stack object retains an authoritative event.
	triggerEvent game.EventKind
	// allowPonderPrefix permits the first spell paragraph of Ponder to lower
	// temporarily. Face lowering rejects it unless the following spell paragraph
	// is the exact typed draw suffix.
	allowPonderPrefix bool
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
	enclosingKind compiler.AbilityKind,
	content compiler.AbilityContent,
	optional bool,
	bodySyntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	ctx := contentCtx{
		text:          bodySyntax.Text,
		span:          bodySyntax.Span,
		optional:      optional,
		content:       content,
		enclosingKind: enclosingKind,
	}
	return lowerContent(cardName, ctx, bodySyntax)
}

func lowerSpellAbilityContent(
	cardName string,
	content compiler.AbilityContent,
	optional bool,
	bodySyntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	ctx := contentCtx{
		text:              bodySyntax.Text,
		span:              bodySyntax.Span,
		optional:          optional,
		content:           content,
		enclosingKind:     compiler.AbilitySpell,
		allowPonderPrefix: true,
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
	enclosingKind compiler.AbilityKind,
	content compiler.AbilityContent,
	optional bool,
	bodySyntax *parser.Ability,
	allowEventPronoun bool,
) (game.AbilityContent, *shared.Diagnostic) {
	ctx := contentCtx{
		text:              bodySyntax.Text,
		span:              bodySyntax.Span,
		optional:          optional,
		content:           content,
		enclosingKind:     enclosingKind,
		sequenceClause:    true,
		allowEventPronoun: allowEventPronoun,
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
		enclosingKind:         compiler.AbilityTriggered,
		triggerCardCountEvent: triggerEvent,
		triggerEvent:          triggerEvent,
	}
	return lowerContent(cardName, ctx, bodySyntax)
}

func lowerContent(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	if content, ok := lowerPonderSequence(ctx); ok {
		return content, nil
	}
	if content, ok := lowerCounterThenNextTurnUpkeepDraws(ctx); ok {
		return content, nil
	}
	if content, ok := lowerControllerPaidEffect(cardName, ctx, syntax); ok {
		return content, nil
	}
	if content, ok := lowerEventPlayerTaxedControllerBenefit(cardName, ctx, syntax); ok {
		return content, nil
	}
	if hasOptionalResolvingEffect(ctx.content.Effects) {
		return lowerOptionalContent(cardName, ctx, syntax)
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
	if content, ok := lowerLinkedSearchUntapSequence(ctx); ok {
		return content, nil
	}
	if len(ctx.content.Effects) > 0 && ctx.content.Effects[0].Kind == compiler.EffectSearch {
		return lowerSearchSpell(ctx)
	}
	if len(ctx.content.Effects) > 1 {
		if len(ctx.content.Effects) == 2 &&
			ctx.content.Effects[0].Kind == compiler.EffectAddMana &&
			isManaSpendRider(&ctx.content.Effects[1]) {
			return lowerManaSpendRiderContent(ctx)
		}
		if ctx.content.Effects[0].Kind == compiler.EffectGainControl ||
			(ctx.content.Effects[0].Kind == compiler.EffectUntap &&
				len(ctx.content.Effects) >= 2 &&
				ctx.content.Effects[1].Kind == compiler.EffectGainControl) {
			return lowerControlSpellSequence(cardName, ctx, syntax)
		}
		return lowerOrderedEffectSequence(cardName, ctx, syntax)
	}
	if len(ctx.content.Effects) == 1 {
		if content, ok := lowerStandaloneReorderLibraryTop(ctx); ok {
			return content, nil
		}
		if ctx.content.Effects[0].RequiresOrderedLowering {
			return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — single effect requires ordered lowering")
		}
		if ctx.content.Effects[0].Kind == compiler.EffectImpulseExile {
			return lowerImpulseExileContent(ctx)
		}
		if ctx.content.Effects[0].Kind == compiler.EffectAddMana {
			return lowerAddManaContent(ctx)
		}
		if content, ok := lowerExileFromHandContent(ctx); ok {
			return content, nil
		}
		// A single effect carrying a per-effect gate condition (e.g. the Addendum
		// "If you cast this spell during your main phase, ...") is lowered through
		// the ordered-sequence path, which applies the supported effect-gate
		// condition to the instruction. The single-effect lowerers reject any
		// condition, so this only adds support and never changes an
		// unconditional single effect.
		if len(ctx.content.Conditions) != 0 {
			gatedCtx := ctx
			gatedCtx.content = contentWithoutConditionSpannedReferences(ctx.content)
			if content, diagnostic := lowerOrderedEffectSequence(cardName, gatedCtx, syntax); diagnostic == nil {
				return content, nil
			}
		}
		return lowerSingleEffectSpell(cardName, ctx, syntax)
	}
	return game.AbilityContent{}, contentDiagnostic(
		ctx,
		"unsupported ability content",
		"the executable source backend does not yet lower this ability content",
	)
}

// contentWithoutConditionSpannedReferences returns a copy of content with every
// reference whose source span lies within a condition clause removed from both
// the content-level reference list and each effect's references. Such a
// reference (e.g. the "this spell" inside "If you cast this spell during your
// main phase, ...") belongs to the gate condition, not to the gated effect, so
// the per-effect lowerers must not mistake it for an effect reference.
func contentWithoutConditionSpannedReferences(content compiler.AbilityContent) compiler.AbilityContent {
	if len(content.Conditions) == 0 {
		return content
	}
	conditionSpans := make([]shared.Span, len(content.Conditions))
	for i := range content.Conditions {
		conditionSpans[i] = content.Conditions[i].Span
	}
	spanned := func(reference compiler.CompiledReference) bool {
		return spanCovered(reference.Span, conditionSpans)
	}
	result := content
	result.References = slices.DeleteFunc(slices.Clone(content.References), spanned)
	result.Effects = slices.Clone(content.Effects)
	for i := range result.Effects {
		result.Effects[i].References = slices.DeleteFunc(slices.Clone(result.Effects[i].References), spanned)
		result.Effects[i].SubjectReferences = slices.DeleteFunc(slices.Clone(result.Effects[i].SubjectReferences), spanned)
	}
	return result
}

// lowerOptionalContent lowers an ability body that carries a resolving optional
// ("you may") effect. Optionality is supported through the ordered effect-sequence
// path for the multi-effect "you may X. If you do, Y" flow and the
// single-optional-effect path for a one-effect "you may X" body, plus the
// dedicated search and removal-then-search shapes. Any other shape fails closed.
func lowerOptionalContent(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	if len(ctx.content.Modes) == 0 &&
		len(ctx.content.Effects) > 1 &&
		ctx.content.Effects[0].Kind != compiler.EffectSearch &&
		!typedManifestDreadSequence(ctx.content) {
		if content, diagnostic := lowerOrderedEffectSequence(cardName, ctx, syntax); diagnostic == nil {
			return content, nil
		}
	}
	if content, ok := lowerSingleOptionalEffect(cardName, ctx, syntax); ok {
		return content, nil
	}
	if content, ok := lowerOptionalHaveEffect(cardName, ctx, syntax); ok {
		return content, nil
	}
	if content, ok := lowerOptionalSearchSpell(ctx); ok {
		return content, nil
	}
	if content, ok := lowerOptionalReferencedControllerSearch(ctx); ok {
		return content, nil
	}
	if content, ok := lowerRemovalThenControllerSearch(cardName, ctx, syntax); ok {
		return content, nil
	}
	if content, ok := lowerOptionalBlinkReturn(cardName, ctx, syntax); ok {
		return content, nil
	}
	return game.AbilityContent{}, contentDiagnostic(
		ctx,
		"unsupported optional effect",
		"the executable source backend does not yet lower optional resolving effects",
	)
}

func lowerImpulseExileContent(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	duration, ok := lowerImpulseExileDuration(effect.Duration)
	if ctx.optional ||
		!effect.Exact ||
		effect.Negated ||
		effect.Context != parser.EffectContextController ||
		!ok ||
		!effect.Amount.Known ||
		effect.Amount.Value < 1 ||
		ctx.content.Unconsumed() {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported impulse exile effect",
			"the executable source backend supports only a fixed-count top-of-library impulse exile with a this-turn or until-end-of-turn play window",
		)
	}
	return game.Mode{Sequence: []game.Instruction{{Primitive: game.ImpulseExile{
		Player:   game.ControllerReference(),
		Amount:   game.Fixed(effect.Amount.Value),
		Duration: duration,
	}}}}.Ability(), nil
}

// lowerImpulseExileDuration maps the supported impulse play windows to their
// runtime durations. Both "this turn" and "until end of turn" grant play
// permission through the end of the current turn; any other window fails closed.
func lowerImpulseExileDuration(duration compiler.DurationKind) (game.EffectDuration, bool) {
	switch duration {
	case compiler.DurationThisTurn:
		return game.DurationThisTurn, true
	case compiler.DurationUntilEndOfTurn:
		return game.DurationUntilEndOfTurn, true
	case compiler.DurationUntilEndOfYourNextTurn:
		return game.DurationUntilEndOfYourNextTurn, true
	default:
		return game.DurationPermanent, false
	}
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
	if ctx.optional || consumed.content.Unconsumed() {
		return unsupported("the executable source backend supports only exact unconditional library-search sequences")
	}
	search := ctx.content.Effects[0]
	if search.Context != parser.EffectContextController {
		return unsupported("the executable source backend supports only searches of your library ending with \"then shuffle\"")
	}
	group, ok := searchGroupSpec(ctx.content.Effects)
	if !ok {
		return unsupported("the executable source backend supports only exact unconditional library-search sequences")
	}

	sequence := []game.Instruction{{Primitive: game.Search{
		Player: game.ControllerReference(),
		Spec:   group.Spec,
		Amount: game.Fixed(group.Amount),
	}}}
	if group.RiderIndex != 0 {
		inst, ok := lowerSearchRider(&ctx.content.Effects[group.RiderIndex])
		if !ok {
			return unsupported("the executable source backend supports only a fixed life-loss or random-discard rider in a library-search sequence")
		}
		sequence = append(sequence, inst)
	}
	if len(ctx.content.Effects) > group.Length {
		if group.Spec.Destination != zone.Library ||
			group.Spec.DestinationPosition != game.SearchPositionTop {
			return unsupported("the executable source backend supports a life-loss rider only on a library-top search")
		}
		life := &ctx.content.Effects[group.Length]
		if !exactControllerLifeLoss(life) {
			return unsupported("the executable source backend supports only a fixed controller life-loss rider")
		}
		sequence = append(sequence, game.Instruction{Primitive: game.LoseLife{
			Player: game.ControllerReference(),
			Amount: game.Fixed(life.Amount.Value),
		}})
	}
	return game.Mode{Sequence: sequence}.Ability(), nil
}

// searchGroupSpec builds the SearchSpec and fixed card count for an exact
// library-search effect group (search, optionally reveal, put, then shuffle),
// independent of which player performs the search. It mirrors the structural
// requirements lowerSearchSpell enforces — a known fixed count, a same-sentence
// span with no delay/duration/negation, a recognized "your library" filter, a
// hand or battlefield destination, and a trailing "then shuffle" — but leaves
// the searching player to the caller so both the controller search ("search
// your library ...") and the affected-permanent's-controller rider ("Its
// controller may search their library ...") share one spec builder. It returns
// ok=false (fail closed) for any group it cannot model exactly.
type searchGroup struct {
	Spec       game.SearchSpec
	Amount     int
	Length     int
	RiderIndex int // index of an optional rider effect lowered after the search; 0 when absent
}

func searchGroupSpec(effects []compiler.CompiledEffect) (searchGroup, bool) {
	shape, ok := exactSearchEffectSequence(effects)
	if !ok {
		return searchGroup{}, false
	}
	search := effects[0]
	if !search.Amount.Known || search.Amount.Value < 1 {
		return searchGroup{}, false
	}
	for i := range shape.length {
		if effects[i].Span != search.Span ||
			effects[i].DelayedTiming != 0 ||
			effects[i].Duration != compiler.DurationNone ||
			effects[i].Negated {
			return searchGroup{}, false
		}
	}
	if search.UnsupportedDetail != "" {
		return searchGroup{}, false
	}
	spec, ok := searchSpecForSelector(search.Selector)
	if !ok {
		return searchGroup{}, false
	}
	spec.SourceZone = zone.Library
	if search.Amount.Value == 1 && spec.IsUnrestricted() {
		spec.FailToFindPolicy = game.SearchMustFindIfAvailable
	}

	if search.SearchSharedSubtype {
		// "that share a land type" correlates the found cards: each must share a
		// land subtype with the others. The runtime enforces it while the cards
		// are chosen, so it is meaningful only for a multi-card search.
		if search.Amount.Value < 2 {
			return searchGroup{}, false
		}
		spec.SharedSubtype = true
	}

	spec.Reveal = shape.reveal
	put := effects[shape.putIndex]
	if shape.top {
		if search.Amount.Value != 1 ||
			search.SearchDestination != parser.EffectDestinationTop ||
			put.Kind != compiler.EffectPut {
			return searchGroup{}, false
		}
		spec.Destination = zone.Library
		spec.DestinationPosition = game.SearchPositionTop
		return searchGroup{Spec: spec, Amount: 1, Length: shape.length}, true
	}
	if split := put.SearchSplit; split.Present {
		// A split-destination put distributes the found cards across two
		// single-card slots, so it requires exactly the two-card "up to two"
		// search. Both slots must be a hand or battlefield destination.
		if search.Amount.Value != 2 ||
			!searchSplitSlotSupported(split.First) ||
			!searchSplitSlotSupported(split.Second) {
			return searchGroup{}, false
		}
		spec.Destination = split.First.ToZone
		spec.EntersTapped = split.First.EntersTapped
		spec.SplitDestination = opt.Val(game.SearchDestination{
			Zone:         split.Second.ToZone,
			EntersTapped: split.Second.EntersTapped,
		})
		return searchGroup{Spec: spec, Amount: search.Amount.Value, Length: shape.length, RiderIndex: shape.riderIndex}, true
	}
	if put.ToZone != zone.Hand && put.ToZone != zone.Battlefield && put.ToZone != zone.Graveyard {
		return searchGroup{}, false
	}
	spec.Destination = put.ToZone
	spec.EntersTapped = put.EntersTapped
	return searchGroup{Spec: spec, Amount: search.Amount.Value, Length: shape.length, RiderIndex: shape.riderIndex}, true
}

// searchSplitSlotSupported reports whether a split-search destination slot names
// a zone the runtime models (hand or battlefield).
func searchSplitSlotSupported(slot parser.SearchSplitSlot) bool {
	return slot.ToZone == zone.Hand || slot.ToZone == zone.Battlefield
}

type searchSequenceShape struct {
	length     int
	putIndex   int
	riderIndex int // index of an optional rider effect between put and shuffle; 0 when absent
	reveal     bool
	top        bool
}

func exactSearchEffectSequence(effects []compiler.CompiledEffect) (searchSequenceShape, bool) {
	if len(effects) < 3 || effects[0].Kind != compiler.EffectSearch {
		return searchSequenceShape{}, false
	}
	if effects[1].Kind == compiler.EffectShuffle && effects[2].Kind == compiler.EffectPut {
		return searchSequenceShape{length: 3, putIndex: 2, top: true}, effects[1].Connection == parser.EffectConnectionThen &&
			effects[2].Connection == parser.EffectConnectionAnd
	}
	if len(effects) == 4 &&
		effects[1].Kind == compiler.EffectReveal &&
		effects[2].Kind == compiler.EffectShuffle &&
		effects[3].Kind == compiler.EffectPut {
		return searchSequenceShape{length: 4, putIndex: 3, reveal: true, top: true}, effects[2].Connection == parser.EffectConnectionThen &&
			effects[3].Connection == parser.EffectConnectionAnd
	}
	return exactSearchPutShuffleSequence(effects)
}

// exactSearchPutShuffleSequence matches the hand/battlefield destination shapes
// "search, [reveal,] put, [rider,] then shuffle." A single optional rider effect
// (a random discard or a fixed controller life loss) may sit between the put and
// the trailing shuffle; lowering validates and lowers it after the search.
func exactSearchPutShuffleSequence(effects []compiler.CompiledEffect) (searchSequenceShape, bool) {
	idx := 1
	reveal := false
	if effects[idx].Kind == compiler.EffectReveal {
		reveal = true
		idx++
	}
	if idx >= len(effects) || effects[idx].Kind != compiler.EffectPut {
		return searchSequenceShape{}, false
	}
	putIndex := idx
	idx++
	riderIndex := 0
	if idx == len(effects)-2 {
		riderIndex = idx
		idx++
	}
	if idx != len(effects)-1 ||
		effects[idx].Kind != compiler.EffectShuffle ||
		effects[idx].Connection != parser.EffectConnectionThen {
		return searchSequenceShape{}, false
	}
	return searchSequenceShape{length: len(effects), putIndex: putIndex, riderIndex: riderIndex, reveal: reveal}, true
}

func exactControllerLifeLoss(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectLose &&
		effect.Context == parser.EffectContextController &&
		effect.LifeObject &&
		effect.Exact &&
		effect.Amount.Known &&
		effect.Amount.Value > 0 &&
		effect.DelayedTiming == 0 &&
		effect.Duration == compiler.DurationNone &&
		!effect.Negated &&
		!effect.Optional
}

// lowerSearchRider lowers a supported rider effect that sits inside a
// library-search sequence (between the put and the trailing shuffle) into the
// instruction that runs after the search primitive. It models a fixed controller
// life loss and a random own-hand discard, failing closed for any other effect.
func lowerSearchRider(rider *compiler.CompiledEffect) (game.Instruction, bool) {
	if exactControllerLifeLoss(rider) {
		return game.Instruction{Primitive: game.LoseLife{
			Player: game.ControllerReference(),
			Amount: game.Fixed(rider.Amount.Value),
		}}, true
	}
	if exactControllerRandomDiscard(rider) {
		return game.Instruction{Primitive: game.Discard{
			Player:   game.ControllerReference(),
			Amount:   game.Fixed(rider.Amount.Value),
			AtRandom: true,
		}}, true
	}
	return game.Instruction{}, false
}

func exactControllerRandomDiscard(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectDiscard &&
		effect.Context == parser.EffectContextController &&
		effect.HandDiscard.Present &&
		effect.HandDiscard.AtRandom &&
		effect.Exact &&
		effect.Amount.Known &&
		effect.Amount.Value > 0 &&
		effect.DelayedTiming == 0 &&
		effect.Duration == compiler.DurationNone &&
		!effect.Negated &&
		!effect.Optional &&
		len(effect.References) == 0
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
		selector.MatchPower ||
		selector.MatchToughness ||
		len(selector.ExcludedTypes()) != 0 ||
		len(selector.ExcludedColors()) != 0 {
		return game.SearchSpec{}, false
	}
	spec.ColorsAny = slices.Clone(selector.ColorsAny())
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
	case compiler.SelectorPlaneswalker:
		spec.CardType = opt.Val(types.Planeswalker)
	case compiler.SelectorPermanent:
		spec.Permanent = true
	default:
		return game.SearchSpec{}, false
	}
	requiredTypesAny := selector.RequiredTypesAny()
	if len(requiredTypesAny) > 0 {
		if selector.Kind == compiler.SelectorPermanent ||
			selector.Kind == compiler.SelectorSpell {
			return game.SearchSpec{}, false
		}
		if len(requiredTypesAny) == 1 {
			// A single required card type reaches lowering only for a plain card
			// selection (the spell types instant and sorcery, which have no
			// dedicated selector kind). It lowers to the singular CardType filter
			// so "a sorcery card" or "an instant card" tutor keeps its type.
			if selector.Kind != compiler.SelectorCard {
				return game.SearchSpec{}, false
			}
			spec.CardType = opt.Val(requiredTypesAny[0])
		} else {
			spec.CardType = opt.V[types.Card]{}
			spec.CardTypesAny = slices.Clone(requiredTypesAny)
		}
	}
	if selector.MatchManaValue {
		// Only the "with mana value N or less" rider is modeled: a fixed upper
		// bound. Every other comparison (exact, "or greater", or an X-derived
		// bound, which reaches lowering as a non-exact clause) fails closed.
		if selector.ManaValue.Op != compare.LessOrEqual {
			return game.SearchSpec{}, false
		}
		spec.MaxManaValue = opt.Val(selector.ManaValue.Value)
	}
	supertypes := selector.Supertypes()
	if len(supertypes) > 1 {
		return game.SearchSpec{}, false
	}
	if len(supertypes) == 1 {
		switch supertypes[0] {
		case types.Basic:
			spec.Supertype = opt.Val(types.Basic)
		case types.Legendary:
			spec.Supertype = opt.Val(types.Legendary)
		default:
			return game.SearchSpec{}, false
		}
	}
	spec.SubtypesAny = slices.Clone(selector.SubtypesAny())
	if selector.BasicLandType {
		if selector.Kind != compiler.SelectorLand || len(spec.SubtypesAny) != 0 ||
			spec.Supertype.Exists {
			return game.SearchSpec{}, false
		}
		spec.SubtypesAny = []types.Sub{
			types.Plains,
			types.Island,
			types.Swamp,
			types.Mountain,
			types.Forest,
		}
	}
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
	if content, diagnostic, handled := lowerPlayerRuleOrPhaseEffect(ctx); handled {
		return content, diagnostic
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

// lowerReferencedPermanentEffect lowers a no-target single effect whose object is
// the source or a singular back-reference. It covers destroy, exile, tap, untap,
// sacrifice, and return-to-hand.
func lowerReferencedPermanentEffect(ctx contentCtx) (game.AbilityContent, bool) {
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
	hasDirectObject := false
	for _, ref := range ctx.content.References {
		if ref.Binding != compiler.ReferenceBindingEventPermanent &&
			ref.Binding != compiler.ReferenceBindingTarget &&
			ref.Binding != compiler.ReferenceBindingSource {
			return game.AbilityContent{}, false
		}
		switch ref.Kind {
		case compiler.ReferencePronoun:
			// "it" names the object directly; the possessive "its" only
			// qualifies a destination ("its owner's hand"), so a body with only
			// "its" carries no direct object and is rejected below.
			if ref.Pronoun != compiler.ReferencePronounIt &&
				ref.Pronoun != compiler.ReferencePronounIts {
				return game.AbilityContent{}, false
			}
			hasDirectObject = hasDirectObject || ref.Pronoun == compiler.ReferencePronounIt
		case compiler.ReferenceThatObject:
			hasDirectObject = true
		case compiler.ReferenceThisObject, compiler.ReferenceSelfName:
			if ref.Binding != compiler.ReferenceBindingSource {
				return game.AbilityContent{}, false
			}
			if ctx.content.Effects[0].Kind != compiler.EffectTap &&
				ctx.content.Effects[0].Kind != compiler.EffectUntap {
				return game.AbilityContent{}, false
			}
			hasDirectObject = true
		default:
			return game.AbilityContent{}, false
		}
	}
	if !hasDirectObject {
		return game.AbilityContent{}, false
	}
	consumed := ctx
	consumed.content.References = nil
	if consumed.content.Unconsumed() {
		return game.AbilityContent{}, false
	}
	object, ok := lowerObjectReference(ctx.content.References[0], referenceLoweringContext{
		AllowEvent:  true,
		AllowSource: true,
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

// lowerDealDamageSpell dispatches a single deal-damage effect to the matching
// damage lowerer, trying the more specific shapes (divided, inherited/source
// power, "each of N targets") before falling back to group and fixed damage so
// the broadest single-target path and its diagnostic stay last.
func lowerDealDamageSpell(cardName string, ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	if ctx.content.Effects[0].Divided {
		return lowerDividedDamageSpell(ctx)
	}
	if ctx.content.Effects[0].DamageRecipientReference == parser.DamageRecipientReferenceYou {
		return lowerControllerDamageSpell(ctx)
	}
	if content, ok := lowerInheritedPowerDamageSpell(ctx); ok {
		return content, nil
	}
	if content, ok := lowerSourcePowerGroupDamageSpell(ctx); ok {
		return content, nil
	}
	if content, ok := lowerSourcePowerDamageSpell(ctx); ok {
		return content, nil
	}
	if content, ok := lowerEachOfTargetsDamageSpell(ctx); ok {
		return content, nil
	}
	if content, ok := lowerEachSourceDamageSpell(ctx); ok {
		return content, nil
	}
	if ctx.content.Effects[0].HasSecondTargetDamageRider {
		return lowerTwoTargetDamageSpell(cardName, ctx)
	}
	if len(ctx.content.Targets) == 0 {
		return lowerGroupDamageSpell(cardName, ctx)
	}
	return lowerFixedDamageSpell(cardName, ctx)
}

func lowerReturnSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	if content, ok := lowerSelfCardGraveyardReturn(ctx); ok {
		return content, nil
	}
	if content, ok := lowerTargetedGraveyardReturn(ctx); ok {
		return content, nil
	}
	if content, ok := lowerChosenCardGraveyardReturn(ctx); ok {
		return content, nil
	}
	if content, ok := lowerMassGraveyardReturn(ctx); ok {
		return content, nil
	}
	if group, ok := exactMassBounceGroup(ctx); ok {
		return game.Mode{
			Sequence: []game.Instruction{{
				Primitive: game.Bounce{Group: group},
			}},
		}.Ability(), nil
	}
	if content, ok := lowerMultiTargetBounceSpell(ctx); ok {
		return content, nil
	}
	if content, ok := lowerDualTargetBounceSpell(ctx); ok {
		return content, nil
	}
	if content, ok := lowerControlledBounceSpell(ctx); ok {
		return content, nil
	}
	if content, ok := lowerSpellBounce(ctx); ok {
		return content, nil
	}
	return lowerFixedBounceSpell(ctx)
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
	if content, ok := lowerReferencedPermanentEffect(ctx); ok {
		return content, nil
	}
	switch ctx.content.Effects[0].Kind {
	case compiler.EffectDealDamage:
		return lowerDealDamageSpell(cardName, ctx)
	case compiler.EffectCantBeBlocked:
		return lowerCantBeBlockedSpell(ctx)
	case compiler.EffectDraw:
		return lowerFixedDrawSpell(ctx, syntax)
	case compiler.EffectDestroy:
		return lowerFixedDestroySpell(ctx)
	case compiler.EffectGain:
		if len(ctx.content.Keywords) != 0 &&
			temporaryKeywordDuration(ctx.content.Effects[0].Duration) {
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
		if len(ctx.content.Keywords) != 0 &&
			temporaryKeywordDuration(ctx.content.Effects[0].Duration) {
			return lowerTemporaryKeywordLossSpell(ctx)
		}
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
		if ctx.content.Effects[0].DiscardEntireHand {
			return lowerDiscardEntireHandSpell(ctx)
		}
		atRandom := ctx.content.Effects[0].HandDiscard.AtRandom
		return lowerFixedCardCountPlayerSpell(
			ctx, syntax, "discard", "discards", false, func(amount game.Quantity, player game.PlayerReference) game.Primitive {
				return game.Discard{Amount: amount, Player: player, AtRandom: atRandom}
			}, func(amount game.Quantity, group game.PlayerGroupReference) game.Primitive {
				return game.Discard{Amount: amount, PlayerGroup: group, AtRandom: atRandom}
			},
		)
	case compiler.EffectMill:
		return lowerFixedCardCountPlayerSpell(
			ctx, syntax, "mill", "mills", true, func(amount game.Quantity, player game.PlayerReference) game.Primitive {
				return game.Mill{Amount: amount, Player: player}
			}, func(amount game.Quantity, group game.PlayerGroupReference) game.Primitive {
				return game.Mill{Amount: amount, PlayerGroup: group}
			},
		)
	case compiler.EffectTap:
		return lowerMassOrSinglePermanentSpell(ctx, "Tap", func(group game.GroupReference) game.Primitive {
			return game.Tap{Group: group}
		}, func(object game.ObjectReference) game.Primitive {
			return game.Tap{Object: object}
		})
	case compiler.EffectUntap:
		return lowerUntapSpell(ctx)
	case compiler.EffectExile:
		return lowerFixedExileSpell(ctx)
	case compiler.EffectReturn:
		return lowerReturnSpell(ctx)
	case compiler.EffectPut:
		return lowerPutEffectSpell(ctx)
	case compiler.EffectMoveCounters:
		return lowerMoveCountersSpell(ctx)
	case compiler.EffectModifyPT:
		return lowerFixedModifyPTSpell(ctx, syntax)
	case compiler.EffectDouble:
		return lowerDoublePTSpell(ctx)
	case compiler.EffectCounter:
		return lowerCounterSpell(ctx)
	case compiler.EffectChooseNewTargets:
		return lowerChooseNewTargetsSpell(ctx)
	case compiler.EffectSacrifice:
		return lowerSacrificeSpell(ctx)
	case compiler.EffectCreate:
		return lowerCreateTokenSpell(ctx)
	case compiler.EffectCast:
		return lowerCastForFreeSpell(ctx)
	case compiler.EffectAttach:
		return lowerAttachSpell(ctx)
	case compiler.EffectWinGame:
		return lowerWinGameSpell(ctx)
	case compiler.EffectMassReanimationExchange:
		return lowerMassReanimationExchangeSpell(ctx)
	case compiler.EffectPunisherLoseLife:
		return lowerPunisherLoseLifeSpell(ctx)
	case compiler.EffectRepeatProcess:
		return lowerRepeatProcessSpell(cardName, ctx, syntax)
	case compiler.EffectPreventDamage:
		return lowerPreventDamageSpell(ctx)
	default:
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported ability content",
			"the executable source backend does not yet lower this ability content",
		)
	}
}

// lowerReturnEffectSpell lowers EffectReturn bodies, trying each supported
// graveyard-return and bounce shape in turn before the fixed-bounce fallback.
func temporaryKeywordDuration(duration compiler.DurationKind) bool {
	return duration == compiler.DurationUntilEndOfTurn ||
		duration == compiler.DurationUntilYourNextTurn
}
