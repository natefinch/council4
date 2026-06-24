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
	// triggerOneOrMore reports whether the enclosing trigger coalesces its
	// simultaneous batch into a single trigger ("Whenever one or more ..."). It
	// gates the batch reanimation of the triggering cards ("put them onto the
	// battlefield") so the plural "them" resolves to the whole batch rather than
	// a single event card.
	triggerOneOrMore bool
	// triggerToZone is the destination zone of the enclosing zone-change
	// trigger, or zone.None outside one. It confirms the triggering cards rest
	// in a graveyard before a batch reanimation recurses them.
	triggerToZone zone.Type
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
	pattern game.TriggerPattern,
) (game.AbilityContent, *shared.Diagnostic) {
	ctx := contentCtx{
		text:                  bodySyntax.Text,
		span:                  bodySyntax.Span,
		optional:              optional,
		content:               content,
		enclosingKind:         compiler.AbilityTriggered,
		triggerCardCountEvent: pattern.Event,
		triggerEvent:          pattern.Event,
		triggerOneOrMore:      pattern.OneOrMore,
		triggerToZone:         triggerPatternToZone(pattern),
	}
	return lowerContent(cardName, ctx, bodySyntax)
}

// triggerPatternToZone reports the destination zone a zone-change trigger
// matches, or zone.None when the pattern does not constrain its destination.
func triggerPatternToZone(pattern game.TriggerPattern) zone.Type {
	if pattern.MatchToZone {
		return pattern.ToZone
	}
	return zone.None
}

func lowerContent(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic) {
	if syntax != nil && syntax.CoinFlip != nil {
		// A recognized coin flip must lower through its dedicated path, which
		// gates every branch effect on the flip result. If that path fails
		// closed (an unsupported branch effect, or a targeted branch), the
		// whole ability fails closed rather than falling through to generic
		// lowering, which would silently drop the flip and emit the branch
		// effects ungated.
		if content, ok := lowerCoinFlipSequence(cardName, ctx, syntax); ok {
			return content, nil
		}
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — coin flip branch not lowered")
	}
	if syntax != nil && syntax.Vote != nil {
		// A recognized vote must lower through its dedicated path, which gates
		// every arm effect on the vote tally. If that path fails closed (an
		// unsupported arm effect, or a targeted arm), the whole ability fails
		// closed rather than falling through to generic lowering, which would
		// silently drop the vote and emit the arm effects ungated.
		if content, ok := lowerVoteSequence(cardName, ctx, syntax); ok {
			return content, nil
		}
		return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — vote arm not lowered")
	}
	if content, ok := lowerPonderSequence(ctx); ok {
		return content, nil
	}
	if content, ok := lowerCounterThenNextTurnUpkeepDraws(ctx); ok {
		return content, nil
	}
	if content, ok := lowerControllerPaidEffect(cardName, ctx, syntax); ok {
		return content, nil
	}
	if content, ok := lowerOptionalPaidBenefit(cardName, ctx, syntax); ok {
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
	if content, ok := lowerEventCardBatchReanimation(ctx); ok {
		return content, nil
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
		if content, diagnostic, handled := lowerOrAlternativeModal(cardName, ctx, syntax); handled {
			return content, diagnostic
		}
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
		if content, ok := lowerAesirExileGraveyardScaledGain(ctx); ok {
			return content, nil
		}
		if content, ok := lowerReturnLinkedExiledPartialContent(ctx); ok {
			return content, nil
		}
		if content, ok := lowerDestroyForEachPlayerTokenChainContent(ctx); ok {
			return content, nil
		}
		return lowerOrderedEffectSequence(cardName, ctx, syntax)
	}
	if len(ctx.content.Effects) == 1 {
		if content, ok := lowerNextCastEntersWithCountersReplacement(ctx); ok {
			return content, nil
		}
		if content, ok := lowerStandaloneReorderLibraryTop(ctx); ok {
			return content, nil
		}
		if content, ok := lowerTrailingBackReferenceExile(ctx); ok {
			return content, nil
		}
		if content, ok := lowerExileUntilLeavesContent(ctx); ok {
			return content, nil
		}
		if content, ok := lowerExileForEachPlayerUntilLeavesContent(ctx); ok {
			return content, nil
		}
		if content, ok := lowerReturnExiledCardContent(ctx); ok {
			return content, nil
		}
		if content, ok := lowerExileEntireHandContent(ctx); ok {
			return content, nil
		}
		if content, ok := lowerReturnExiledCardsToHandContent(ctx); ok {
			return content, nil
		}
		if content, ok := lowerBottomLinkedExiledCardsContent(ctx); ok {
			return content, nil
		}
		if content, ok := lowerAesirCounterFromExiledCard(ctx); ok {
			return content, nil
		}
		if content, ok := lowerAesirReturnSourceAndExiledCard(ctx); ok {
			return content, nil
		}
		if content, ok := lowerStandaloneStunEffect(ctx); ok {
			return content, nil
		}
		if content, ok := lowerStandaloneSourceStunEffect(ctx); ok {
			return content, nil
		}
		if ctx.content.Effects[0].RequiresOrderedLowering {
			return game.AbilityContent{}, unsupportedEffectSequenceDiagnostic(ctx, "structural — single effect requires ordered lowering")
		}
		switch ctx.content.Effects[0].Kind {
		case compiler.EffectImpulseExile:
			return lowerImpulseExileContent(ctx)
		case compiler.EffectAddMana:
			return lowerAddManaContent(ctx)
		case compiler.EffectBecomeCopy:
			return lowerBecomeCopyContent(ctx)
		case compiler.EffectBecomeType:
			return lowerBecomeTypeContent(ctx)
		case compiler.EffectPolymorph:
			return lowerPolymorphContent(ctx)
		case compiler.EffectSetBasePT:
			return lowerSetBasePTContent(ctx)
		case compiler.EffectDelayedTrigger:
			return lowerDelayedTriggerContent(ctx)
		case compiler.EffectCreateEmblem:
			return lowerCreateEmblemContent(ctx)
		default:
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
	if content, ok := lowerConditionalDestinationPlace(ctx); ok {
		return content, nil
	}
	if content, ok := lowerKinshipReveal(cardName, ctx, syntax); ok {
		return content, nil
	}
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
		(!effect.Amount.Known && !effect.Amount.VariableX) ||
		(effect.Amount.Known && effect.Amount.Value < 1) ||
		ctx.content.Unconsumed() {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported impulse exile effect",
			"the executable source backend supports only a fixed-count top-of-library impulse exile with a this-turn or until-end-of-turn play window",
		)
	}
	amount := game.Fixed(effect.Amount.Value)
	if effect.Amount.VariableX {
		amount = game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	}
	return game.Mode{Sequence: []game.Instruction{{Primitive: game.ImpulseExile{
		Player:   game.ControllerReference(),
		Amount:   amount,
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

func hasOptionalPaymentResolvingEffect(effects []compiler.CompiledEffect) bool {
	for i := range effects {
		if effects[i].Payment.Form == parser.EffectPaymentFormMayPayThenIfDo {
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
	// A trailing "Shuffle this card into its owner's library." tail (Green Sun's
	// Zenith) is the resolving spell shuffling itself back in. Strip it from the
	// search-sequence analysis and re-append it as a source-spell shuffle
	// instruction after the search resolves.
	appendSelfShuffle := false
	if n := len(ctx.content.Effects); n >= 2 && isExactSourceSpellShuffleIntoLibrary(&ctx.content.Effects[n-1]) {
		appendSelfShuffle = true
		shuffleSpan := ctx.content.Effects[n-1].ClauseSpan
		ctx.content.Effects = ctx.content.Effects[:n-1]
		ctx.content.References = referencesOutsideSpan(ctx.content.References, shuffleSpan)
	}
	// The search subject is either the controller ("search your library ...") or
	// a single target player ("target player searches their library ..."). The
	// target-player form contributes an ability target spec and resolves the
	// searcher to that target; every other subject fails closed.
	search := ctx.content.Effects[0]
	subject, ok := searchSearcher(ctx, &search)
	if !ok {
		return unsupported("the executable source backend supports only searches of your library or a single target player's library ending with \"then shuffle\"")
	}
	searcher, searcherGroup, searchTargets := subject.Player, subject.Group, subject.Targets
	// Search is one runtime primitive, but each reference still binds to the
	// prior semantic search/reveal instruction that produced the found card, or
	// to the searching player(s) ("their library").
	targetSearcher := len(searchTargets) != 0
	groupSearcher := searcherGroup.Kind != game.PlayerGroupReferenceNone
	for _, ref := range ctx.content.References {
		if ref.Binding == compiler.ReferenceBindingPriorInstructionResult {
			continue
		}
		if targetSearcher && ref.Binding == compiler.ReferenceBindingTarget && isPlayerPronoun(ref.Pronoun) {
			continue
		}
		// "Each player searches their library ..." — the "their" possessive
		// refers to each searching player and is realized by the all-players
		// group searcher, so no per-reference lowering is required.
		if groupSearcher && isPlayerPronoun(ref.Pronoun) {
			continue
		}
		return unsupported("unexpected non-result reference in search effect")
	}
	consumed := ctx
	consumed.content.References = nil
	if targetSearcher {
		consumed.content.Targets = nil
	}
	if ctx.optional || consumed.content.Unconsumed() {
		return unsupported("the executable source backend supports only exact unconditional library-search sequences")
	}
	group, ok := searchGroupSpec(ctx.content.Effects)
	if !ok {
		return unsupported("the executable source backend supports only exact unconditional library-search sequences")
	}
	controller, controllerTargets, ok := searchController(search.SearchControl, searchTargets, group.Spec)
	if !ok {
		return unsupported("the executable source backend supports the \"under target player's control\" rider only on a single-target battlefield search with no other target")
	}
	searchTargets = append(searchTargets, controllerTargets...)

	if searcherGroup.Kind != game.PlayerGroupReferenceNone && controller.Exists {
		return unsupported("the executable source backend does not support the \"under target player's control\" rider on an each-player library search")
	}
	sequence := []game.Instruction{{Primitive: game.Search{
		Player:      searcher,
		PlayerGroup: searcherGroup,
		Spec:        group.Spec,
		Amount:      game.Fixed(group.Amount),
		Controller:  controller,
	}}}
	if group.RiderIndex != 0 {
		inst, ok := lowerSearchRider(&ctx.content.Effects[group.RiderIndex])
		if !ok {
			return unsupported("the executable source backend supports only a fixed life-loss or random-discard rider in a library-search sequence")
		}
		sequence = append(sequence, inst)
	}
	// Effects after the search group (its trailing "then shuffle") are riders
	// that resolve once the search completes — Grim Tutor's "You lose 3 life.",
	// Environmental Sciences' "You gain 2 life."
	topSearch := group.Spec.Destination == zone.Library &&
		group.Spec.DestinationPosition == game.SearchPositionTop
	for i := group.Length; i < len(ctx.content.Effects); i++ {
		effect := &ctx.content.Effects[i]
		// Top-of-library tutors keep their tighter contract: only a fixed
		// controller life-loss rider (Vampiric Tutor, Imperial Seal).
		if topSearch {
			if !exactControllerLifeLoss(effect) {
				return unsupported("the executable source backend supports a fixed life-loss rider only on a library-top search")
			}
			sequence = append(sequence, game.Instruction{Primitive: game.LoseLife{
				Player: game.ControllerReference(),
				Amount: game.Fixed(effect.Amount.Value),
			}})
			continue
		}
		inst, ok := lowerSearchRider(effect)
		if !ok {
			return unsupported("the executable source backend supports only fixed controller life-change or random-discard riders after a library-search sequence")
		}
		sequence = append(sequence, inst)
	}
	if appendSelfShuffle {
		sequence = append(sequence, game.Instruction{Primitive: game.ShuffleSpellIntoLibrary{}})
	}
	return game.Mode{Targets: searchTargets, Sequence: sequence}.Ability(), nil
}

// searchSubject captures the player or player group performing a library search
// and any ability target specs that searcher reference requires.
type searchSubject struct {
	Player  game.PlayerReference
	Group   game.PlayerGroupReference
	Targets []game.TargetSpec
}

// searchSearcher determines the player performing a library search and any
// ability target specs that searcher reference requires. It supports the
// controller subject ("search your library ...") , a single target player
// subject ("target player searches their library ..."), resolving the latter to
// TargetPlayerReference(0) with the matching player target spec, and the
// each-player subject ("each player searches their library ..."), resolving to
// the all-players group so every player searches their own library. Every other
// subject — a referenced object's controller, an unsupported target shape —
// fails closed so lowering never invents a searcher.
func searchSearcher(ctx contentCtx, search *compiler.CompiledEffect) (searchSubject, bool) {
	switch search.Context {
	case parser.EffectContextController:
		if len(ctx.content.Targets) != 0 {
			return searchSubject{}, false
		}
		return searchSubject{Player: game.ControllerReference()}, true
	case parser.EffectContextTarget:
		if len(ctx.content.Targets) != 1 {
			return searchSubject{}, false
		}
		spec, ok := playerTargetSpec(ctx.content.Targets[0])
		if !ok {
			return searchSubject{}, false
		}
		return searchSubject{Player: game.TargetPlayerReference(0), Targets: []game.TargetSpec{spec}}, true
	case parser.EffectContextEachPlayer:
		if len(ctx.content.Targets) != 0 {
			return searchSubject{}, false
		}
		return searchSubject{Group: game.AllPlayersReference()}, true
	default:
		return searchSubject{}, false
	}
}

// isPlayerPronoun reports whether a reference pronoun names a player (they /
// their / them), distinguishing the searching target player's "their library"
// reference from the found card's "it" result reference.
func isPlayerPronoun(pronoun compiler.ReferencePronounKind) bool {
	switch pronoun {
	case compiler.ReferencePronounThey,
		compiler.ReferencePronounTheir,
		compiler.ReferencePronounThem:
		return true
	default:
		return false
	}
}

// searchController resolves the player a found permanent enters under and any
// ability target spec that controller reference requires. With no rider the
// found card enters under the searching player's control, so no Controller
// reference is set. The "under target player's/opponent's control" rider
// (Yavimaya Dryad) routes the found permanent to a target player, adding that
// player target spec and binding Search.Controller to it. It is supported only
// on a single-card battlefield search whose ability has no other target — the
// controller target is the ability's sole target — so the searcher-is-target
// and enters-under-target forms never share a target index. Every other pairing
// fails closed.
func searchController(control parser.SearchControlRider, existingTargets []game.TargetSpec, spec game.SearchSpec) (opt.V[game.PlayerReference], []game.TargetSpec, bool) {
	if control == parser.SearchControlRiderNone {
		return opt.V[game.PlayerReference]{}, nil, true
	}
	if len(existingTargets) != 0 ||
		spec.Destination != zone.Battlefield ||
		spec.SplitDestination.Exists {
		return opt.V[game.PlayerReference]{}, nil, false
	}
	switch control {
	case parser.SearchControlRiderTargetPlayer:
		return opt.Val(game.TargetPlayerReference(0)), []game.TargetSpec{controllerPlayerTargetSpec(false)}, true
	case parser.SearchControlRiderTargetOpponent:
		return opt.Val(game.TargetPlayerReference(0)), []game.TargetSpec{controllerPlayerTargetSpec(true)}, true
	default:
		return opt.V[game.PlayerReference]{}, nil, false
	}
}

// controllerPlayerTargetSpec builds the player target spec a found permanent's
// "under target player's/opponent's control" rider chooses as the permanent's
// new controller.
func controllerPlayerTargetSpec(opponent bool) game.TargetSpec {
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: "target player",
		Allow:      game.TargetAllowPlayer,
	}
	if opponent {
		spec.Constraint = "target opponent"
		spec.Predicate = game.TargetPredicate{Player: game.PlayerOpponent}
	}
	return spec
}

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
// "search, [reveal,] put, [rider,] then shuffle[. <trailing rider>]." A single
// optional rider effect (a random discard or a fixed controller life loss) may
// sit between the put and the trailing shuffle; lowering validates and lowers it
// after the search. The group ends at the trailing shuffle, so any effects after
// it (e.g. Grim Tutor's "You lose 3 life.") are left for the caller to lower as
// post-search riders rather than being folded into the search shape.
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
	if idx < len(effects) && effects[idx].Kind != compiler.EffectShuffle {
		riderIndex = idx
		idx++
	}
	if idx >= len(effects) ||
		effects[idx].Kind != compiler.EffectShuffle ||
		effects[idx].Connection != parser.EffectConnectionThen {
		return searchSequenceShape{}, false
	}
	return searchSequenceShape{length: idx + 1, putIndex: putIndex, riderIndex: riderIndex, reveal: reveal}, true
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

func exactControllerLifeGain(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectGain &&
		effect.Context == parser.EffectContextController &&
		effect.LifeObject &&
		effect.Exact &&
		effect.Amount.Known &&
		effect.Amount.Value > 0 &&
		effect.DelayedTiming == 0 &&
		effect.Duration == compiler.DurationNone &&
		!effect.Negated &&
		!effect.Optional &&
		len(effect.References) == 0
}

// lowerSearchRider lowers a supported rider effect that resolves as part of a
// library-search sequence — one sitting between the put and the trailing shuffle,
// or one trailing the shuffle as its own sentence — into the instruction that
// runs after the search primitive. It models a fixed controller life loss or
// gain and a random own-hand discard, failing closed for any other effect.
func lowerSearchRider(rider *compiler.CompiledEffect) (game.Instruction, bool) {
	if exactControllerLifeLoss(rider) {
		return game.Instruction{Primitive: game.LoseLife{
			Player: game.ControllerReference(),
			Amount: game.Fixed(rider.Amount.Value),
		}}, true
	}
	if exactControllerLifeGain(rider) {
		return game.Instruction{Primitive: game.GainLife{
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
		len(selector.ExcludedTypes()) != 0 ||
		len(selector.ExcludedColors()) != 0 {
		return game.SearchSpec{}, false
	}
	var filter game.Selection
	if len(selector.Alternatives) > 0 {
		return searchSpecForAlternatives(selector)
	}
	filter.ColorsAny = slices.Clone(selector.ColorsAny())
	filter.Colorless = selector.Colorless
	spec.Name = selector.RequiredName
	switch selector.Kind {
	case compiler.SelectorCard:
	case compiler.SelectorLand:
		filter.RequiredTypes = []types.Card{types.Land}
	case compiler.SelectorCreature:
		filter.RequiredTypes = []types.Card{types.Creature}
	case compiler.SelectorArtifact:
		filter.RequiredTypes = []types.Card{types.Artifact}
	case compiler.SelectorEnchantment:
		filter.RequiredTypes = []types.Card{types.Enchantment}
	case compiler.SelectorPlaneswalker:
		filter.RequiredTypes = []types.Card{types.Planeswalker}
	case compiler.SelectorPermanent:
		filter.RequirePermanentCard = true
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
			// dedicated selector kind). It lowers to the singular RequiredTypes
			// filter so "a sorcery card" or "an instant card" tutor keeps its type.
			if selector.Kind != compiler.SelectorCard {
				return game.SearchSpec{}, false
			}
			filter.RequiredTypes = []types.Card{requiredTypesAny[0]}
		} else {
			filter.RequiredTypes = nil
			filter.RequiredTypesAny = slices.Clone(requiredTypesAny)
		}
	}
	if selector.MatchManaValue {
		// "with mana value N or less" is a fixed upper bound; "with mana value X
		// or less" binds the upper bound to the spell's chosen {X}, resolved as
		// the search runs. Every other comparison (exact, "or greater") fails
		// closed.
		if selector.ManaValueX {
			spec.MaxManaValueFromX = true
		} else {
			if selector.ManaValue.Op != compare.LessOrEqual {
				return game.SearchSpec{}, false
			}
			filter.ManaValue = opt.Val(selector.ManaValue)
		}
	}
	if selector.MatchPower {
		switch selector.Power.Op {
		case compare.LessOrEqual, compare.GreaterOrEqual:
			filter.Power = opt.Val(selector.Power)
		default:
			return game.SearchSpec{}, false
		}
	}
	if selector.MatchToughness {
		switch selector.Toughness.Op {
		case compare.LessOrEqual, compare.GreaterOrEqual:
			filter.Toughness = opt.Val(selector.Toughness)
		default:
			return game.SearchSpec{}, false
		}
	}
	supertypes := selector.Supertypes()
	if len(supertypes) > 1 {
		return game.SearchSpec{}, false
	}
	if len(supertypes) == 1 {
		switch supertypes[0] {
		case types.Basic:
			filter.Supertypes = []types.Super{types.Basic}
		case types.Legendary:
			filter.Supertypes = []types.Super{types.Legendary}
		default:
			return game.SearchSpec{}, false
		}
	}
	filter.SubtypesAny = slices.Clone(selector.SubtypesAny())
	if selector.BasicLandType {
		if selector.Kind != compiler.SelectorLand || len(filter.SubtypesAny) != 0 ||
			len(filter.Supertypes) != 0 {
			return game.SearchSpec{}, false
		}
		filter.SubtypesAny = []types.Sub{
			types.Plains,
			types.Island,
			types.Swamp,
			types.Mountain,
			types.Forest,
		}
	}
	spec.Filter = filter
	return spec, true
}

// searchSpecForAlternatives lowers a disjunctive search selector (one whose
// sides parsed into Alternatives) into a SearchSpec whose filter is a
// Selection.AnyOf of the per-side filters. The parent selector carries only the
// alternatives, so it must bear no flat type, supertype, subtype, color, name,
// or numeric constraint that AnyOf could not preserve, and each side must lower
// to a plain filter with no name or X-bounded mana value. It fails closed
// otherwise so an unrepresentable disjunction is never silently dropped.
func searchSpecForAlternatives(selector compiler.CompiledSelector) (game.SearchSpec, bool) {
	if selector.Kind != compiler.SelectorUnknown ||
		len(selector.RequiredTypesAny()) != 0 ||
		len(selector.Supertypes()) != 0 ||
		len(selector.ExcludedSupertypes()) != 0 ||
		len(selector.SubtypesAny()) != 0 ||
		len(selector.ExcludedSubtypes()) != 0 ||
		len(selector.ColorsAny()) != 0 ||
		selector.Colorless ||
		selector.RequiredName != "" ||
		selector.BasicLandType ||
		selector.MatchManaValue ||
		selector.MatchPower ||
		selector.MatchToughness {
		return game.SearchSpec{}, false
	}
	var spec game.SearchSpec
	var filter game.Selection
	for i := range selector.Alternatives {
		altSpec, ok := searchSpecForSelector(selector.Alternatives[i])
		if !ok {
			return game.SearchSpec{}, false
		}
		if altSpec.Name != "" || altSpec.MaxManaValueFromX {
			return game.SearchSpec{}, false
		}
		filter.AnyOf = append(filter.AnyOf, altSpec.Filter)
	}
	spec.Filter = filter
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
	if ctx.content.Effects[0].DamageRecipientReference == parser.DamageRecipientReferenceThatPlayer {
		return lowerEventPlayerDamageSpell(ctx)
	}
	if content, ok := lowerInheritedPowerDamageSpell(ctx); ok {
		return content, nil
	}
	if content, ok := lowerInheritedPowerGroupDamageSpell(ctx); ok {
		return content, nil
	}
	if content, ok := lowerSourcePowerGroupDamageSpell(ctx); ok {
		return content, nil
	}
	if content, ok := lowerSourcePowerDamageSpell(ctx); ok {
		return content, nil
	}
	if content, ok := lowerEventPowerGroupDamageSpell(ctx); ok {
		return content, nil
	}
	if content, ok := lowerEachOfTargetsDamageSpell(ctx); ok {
		return content, nil
	}
	if content, ok := lowerEachSelfPowerDamageSpell(ctx); ok {
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
	if content, ok := lowerTotalManaValueGraveyardReanimation(ctx); ok {
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
	case compiler.EffectCantBlock:
		return lowerCantBlockSpell(ctx)
	case compiler.EffectDraw:
		return lowerFixedDrawSpell(ctx, syntax)
	case compiler.EffectDestroy:
		return lowerFixedDestroySpell(ctx)
	case compiler.EffectGain:
		return lowerGainSpellEffect(ctx)
	case compiler.EffectGainControl:
		return lowerSingleControlSpell(ctx)
	case compiler.EffectLose:
		return lowerLoseSpellEffect(ctx)
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
	case compiler.EffectGainPlayerCounter:
		return lowerGainPlayerCounterSpell(ctx, syntax)
	case compiler.EffectBecomeMonarch:
		return lowerBecomeMonarchSpell(ctx)
	case compiler.EffectRingTempts:
		return lowerRingTemptsSpell(ctx)
	case compiler.EffectAmass:
		return lowerAmassContent(ctx, syntax)
	case compiler.EffectRenown:
		return lowerRenownContent(ctx, syntax)
	case compiler.EffectAdapt:
		return lowerAdaptContent(ctx, syntax)
	case compiler.EffectConnive:
		return lowerConniveContent(ctx)
	case compiler.EffectProliferate:
		return lowerExactPrimitiveSpell(ctx, syntax, "proliferate", func(amount game.Quantity) game.Primitive {
			return game.Proliferate{Amount: amount}
		})
	case compiler.EffectExplore:
		return lowerExploreSpell(ctx)
	case compiler.EffectManifest, compiler.EffectManifestDread:
		return lowerManifestSpell(ctx)
	case compiler.EffectRegenerate:
		return lowerRegenerateSpell(ctx)
	case compiler.EffectFight:
		if len(ctx.content.Targets) == 1 {
			return lowerSourceFightSpell(ctx)
		}
		return lowerFightSpell(ctx)
	case compiler.EffectLookAtHand:
		return lowerLookAtHandSpell(ctx)
	case compiler.EffectDiscard:
		if ctx.content.Effects[0].DiscardEntireHand {
			return lowerDiscardEntireHandSpell(ctx)
		}
		if content, ok := lowerFilteredControllerDiscard(ctx); ok {
			return content, nil
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
	case compiler.EffectTapOrUntap:
		return lowerFixedPermanentTargetSpell(ctx, "Tap or untap", func(object game.ObjectReference) game.Primitive {
			return game.TapOrUntap{Object: object}
		})
	case compiler.EffectUntap:
		return lowerUntapSpell(ctx)
	case compiler.EffectRemoveFromCombat:
		return lowerFixedPermanentTargetSpell(ctx, "remove from combat", func(object game.ObjectReference) game.Primitive {
			return game.RemoveFromCombat{Object: object}
		})
	case compiler.EffectExile:
		if len(ctx.content.Effects) == 1 &&
			ctx.content.Effects[0].CardSource == parser.EffectCardSourceTopOfPlayerLibrary {
			return lowerFixedCardCountPlayerSpell(
				ctx, syntax, "exile", "exiles", false, func(amount game.Quantity, player game.PlayerReference) game.Primitive {
					return game.ExileTopOfLibrary{Amount: amount, Player: player}
				}, func(amount game.Quantity, group game.PlayerGroupReference) game.Primitive {
					return game.ExileTopOfLibrary{Amount: amount, PlayerGroup: group}
				},
			)
		}
		return lowerFixedExileSpell(ctx)
	case compiler.EffectShuffle:
		if content, ok := lowerSourceSpellShuffleIntoLibrary(ctx); ok {
			return content, nil
		}
		if content, ok := lowerControllerGraveyardShuffleIntoLibrary(ctx); ok {
			return content, nil
		}
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported shuffle effect",
			"the executable source backend supports only a source-spell shuffle into its owner's library or a controller graveyard shuffle into library",
		)
	case compiler.EffectReturn:
		return lowerReturnSpell(ctx)
	case compiler.EffectPut:
		return lowerPutEffectSpell(ctx)
	case compiler.EffectMoveCounters:
		return lowerMoveCountersSpell(ctx)
	case compiler.EffectRemoveCounter:
		return lowerRemoveCounterSpell(ctx)
	default:
		if content, diag, ok := lowerImmediateSingleEffectSpellTail(cardName, ctx, syntax); ok {
			return content, diag
		}
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported ability content",
			"the executable source backend does not yet lower this ability content",
		)
	}
}

// lowerImmediateSingleEffectSpellTail handles the remaining single-effect kinds
// that lowerImmediateSingleEffectSpell does not dispatch directly, keeping that
// function's maintainability index within bounds. It returns ok=false for any
// effect kind it does not handle so the caller can emit the generic
// unsupported-content diagnostic.
func lowerImmediateSingleEffectSpellTail(
	cardName string,
	ctx contentCtx,
	syntax *parser.Ability,
) (game.AbilityContent, *shared.Diagnostic, bool) {
	switch ctx.content.Effects[0].Kind {
	case compiler.EffectModifyPT:
		content, diag := lowerFixedModifyPTSpell(ctx, syntax)
		return content, diag, true
	case compiler.EffectDouble:
		if ctx.content.Effects[0].DoubleSourceCounters {
			content, diag := lowerDoubleCountersSpell(ctx)
			return content, diag, true
		}
		content, diag := lowerDoublePTSpell(ctx)
		return content, diag, true
	case compiler.EffectCounter:
		content, diag := lowerCounterSpell(ctx)
		return content, diag, true
	case compiler.EffectCopyStackObject:
		content, diag := lowerCopyStackObjectSpell(ctx)
		return content, diag, true
	case compiler.EffectChooseNewTargets:
		content, diag := lowerChooseNewTargetsSpell(ctx)
		return content, diag, true
	case compiler.EffectChooseCreatureType:
		content, diag := lowerChooseCreatureTypeSpell(ctx)
		return content, diag, true
	case compiler.EffectSacrifice:
		content, diag := lowerSacrificeSpell(ctx)
		return content, diag, true
	case compiler.EffectCreate:
		content, diag := lowerCreateTokenSpell(ctx)
		return content, diag, true
	case compiler.EffectCast:
		if content, diag, ok := lowerCastFromGraveyardPermission(ctx); ok {
			return content, diag, true
		}
		content, diag := lowerCastForFreeSpell(ctx)
		return content, diag, true
	case compiler.EffectAttach:
		content, diag := lowerAttachSpell(ctx)
		return content, diag, true
	case compiler.EffectWinGame:
		content, diag := lowerWinGameSpell(ctx)
		return content, diag, true
	case compiler.EffectLoseGame:
		content, diag := lowerLoseGameSpell(ctx)
		return content, diag, true
	case compiler.EffectMassReanimationExchange:
		content, diag := lowerMassReanimationExchangeSpell(ctx)
		return content, diag, true
	case compiler.EffectPunisherLoseLife:
		content, diag := lowerPunisherLoseLifeSpell(ctx)
		return content, diag, true
	case compiler.EffectRepeatProcess:
		content, diag := lowerRepeatProcessSpell(cardName, ctx, syntax)
		return content, diag, true
	case compiler.EffectPreventDamage:
		content, diag := lowerPreventDamageSpell(ctx)
		return content, diag, true
	default:
		return game.AbilityContent{}, nil, false
	}
}

// lowerGainSpellEffect lowers an EffectGain body: either a temporary keyword
// grant, a life-gain effect, or an unsupported keyword/ability grant.
func lowerGainSpellEffect(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	if len(ctx.content.Keywords) != 0 &&
		temporaryKeywordDuration(ctx.content.Effects[0].Duration) {
		return lowerTemporaryKeywordSpell(ctx)
	}
	if len(ctx.content.Keywords) != 0 {
		if _, ok := permanentKeywordGrantDuration(ctx.content.Effects[0].Duration); ok {
			return lowerPermanentKeywordGrantSpell(ctx)
		}
	}
	if ctx.content.Effects[0].GainGrantedAbility != nil {
		return lowerGainGrantedAbilitySpell(ctx)
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
}

// permanentKeywordGrantDuration maps the compiled duration of a resolving keyword
// grant to its runtime EffectDuration. A no-duration grant lasts as long as the
// subject remains on the battlefield (DurationPermanent); the "for as long as you
// control this <noun>" form expires when the source leaves the controller's
// control. It returns ok=false for any other duration so richer grants stay
// fail-closed.
func permanentKeywordGrantDuration(duration compiler.DurationKind) (game.EffectDuration, bool) {
	switch duration {
	case compiler.DurationNone:
		return game.DurationPermanent, true
	case compiler.DurationForAsLongAsYouControlSource:
		return game.DurationForAsLongAsYouControlSource, true
	default:
		return game.DurationPermanent, false
	}
}

// lowerPermanentKeywordGrantSpell lowers a keyword grant to a referenced object
// ("Return target creature card ... to the battlefield. It gains haste.") or to a
// single targeted permanent ("Target creature you control gains indestructible
// for as long as you control this Saga.") into a game.ApplyContinuous that adds
// the keyword for the grant's lifetime. A no-duration grant persists for as long
// as the object remains on the battlefield (DurationPermanent); the "for as long
// as you control this <noun>" form expires when its controller loses the source.
// "It" binds to the prior target, so the no-duration grant composes with
// reanimation and similar back-referencing sequences. It fails closed for any
// shape other than an exact, non-negated keyword grant with a supported duration
// to a referenced or single targeted permanent.
func lowerPermanentKeywordGrantSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported keyword or ability grant",
			"the executable source backend does not yet lower spells that grant a keyword or quoted ability",
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
	duration, durationOK := permanentKeywordGrantDuration(effect.Duration)
	if len(ctx.content.Effects) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		effect.Kind != compiler.EffectGain ||
		!effect.Exact ||
		(!referencedObject && !targetSubject) ||
		effect.Negated ||
		effect.StaticSubject != compiler.StaticSubjectNone ||
		!durationOK {
		return unsupported()
	}
	keywords, abilities, ok := partitionTemporaryKeywords(ctx.content.Keywords)
	if !ok {
		return unsupported()
	}
	var object game.ObjectReference
	var target opt.V[game.TargetSpec]
	switch {
	case targetSubject:
		spec, ok := permanentTargetSpec(ctx.content.Targets[0])
		if !ok {
			return unsupported()
		}
		target = opt.Val(spec)
		object = game.TargetPermanentReference(0)
	default:
		object, ok = lowerObjectReference(ctx.content.References[0], referenceLoweringContext{AllowTarget: true})
		if !ok {
			return unsupported()
		}
	}
	if effect.KeywordGrantChoice {
		if duration != game.DurationPermanent {
			return unsupported()
		}
		return lowerPermanentKeywordChoiceGrant(keywords, abilities, object, target)
	}
	mode := game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.ApplyContinuous{
				Object: opt.Val(object),
				ContinuousEffects: []game.ContinuousEffect{{
					Layer:        game.LayerAbility,
					AddKeywords:  keywords,
					AddAbilities: abilities,
				}},
				Duration: duration,
			},
		}},
	}
	if target.Exists {
		mode.Targets = []game.TargetSpec{target.Val}
	}
	return mode.Ability(), nil
}

// lowerPermanentKeywordChoiceGrant lowers a disjunctive keyword grant ("that
// creature gains banding, first strike, or trample") into a modal ability whose
// modes each grant one of the listed keywords indefinitely. The controller picks
// exactly one mode at resolution, which realizes the "choose one of the listed
// keywords" semantics with the existing modal machinery. The single grant target
// is shared across every mode. Abilities (such as a granted protection static
// body) are never produced by a choice list, so a non-empty abilities slice is
// rejected as unrepresentable.
func lowerPermanentKeywordChoiceGrant(
	keywords []game.Keyword,
	abilities []game.Ability,
	object game.ObjectReference,
	target opt.V[game.TargetSpec],
) (game.AbilityContent, *shared.Diagnostic) {
	if len(keywords) < 2 || len(abilities) != 0 {
		return game.AbilityContent{}, &shared.Diagnostic{
			Severity: shared.SeverityWarning,
			Summary:  "unsupported keyword choice grant",
			Detail:   "the executable source backend supports only a choice among two or more simple grantable keywords",
		}
	}
	modes := make([]game.Mode, 0, len(keywords))
	for _, keyword := range keywords {
		modes = append(modes, game.Mode{
			Sequence: []game.Instruction{{
				Primitive: game.ApplyContinuous{
					Object: opt.Val(object),
					ContinuousEffects: []game.ContinuousEffect{{
						Layer:       game.LayerAbility,
						AddKeywords: []game.Keyword{keyword},
					}},
					Duration: game.DurationPermanent,
				},
			}},
		})
	}
	content := game.AbilityContent{
		Modes:    modes,
		MinModes: 1,
		MaxModes: 1,
	}
	if target.Exists {
		content.SharedTargets = []game.TargetSpec{target.Val}
	}
	return content, nil
}

// lowerLoseSpellEffect lowers an EffectLose body: either a temporary keyword
// loss, a life-loss effect, or an unsupported keyword/ability loss.
func lowerLoseSpellEffect(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
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
}

// lowerReturnEffectSpell lowers EffectReturn bodies, trying each supported
// graveyard-return and bounce shape in turn before the fixed-bounce fallback.
func temporaryKeywordDuration(duration compiler.DurationKind) bool {
	return duration == compiler.DurationUntilEndOfTurn ||
		duration == compiler.DurationUntilYourNextTurn
}
