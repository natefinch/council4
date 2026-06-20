package parser

import (
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

func emitResolvingSyntax(abilities []Ability) {
	for i := range abilities {
		if recognizeChosenTypeLibraryTopSequence(&abilities[i]) {
			continue
		}
		emitSentenceResolvingSyntax(
			abilities[i].Sentences,
			abilities[i].Atoms,
			abilities[i].ActivationRestrictions,
			abilities[i].TriggerFrequency,
			abilities[i].SourceAbilityCostReduction,
		)
		recognizeControllerOptionalPaymentSequence(&abilities[i])
		recognizeEventPlayerOptionalPaymentSequence(&abilities[i])
		if abilities[i].Modal == nil {
			continue
		}
		for j := range abilities[i].Modal.Options {
			mode := &abilities[i].Modal.Options[j]
			emitSentenceResolvingSyntax(mode.Sentences, mode.Atoms, nil, nil, nil)
			if sentencesHaveImpulseExile(mode.Sentences) {
				mode.SemanticReferences = nil
				mode.ConditionBoundaries = nil
				mode.EventHistoryConditions = nil
				mode.ConditionClauses = nil
				mode.ConditionSegments = nil
			}
		}
	}
}

func sentencesHaveImpulseExile(sentences []Sentence) bool {
	return len(sentences) == 2 &&
		len(sentences[0].Effects) == 1 &&
		sentences[0].Effects[0].Kind == EffectImpulseExile
}

func emitSentenceResolvingSyntax(
	sentences []Sentence,
	atoms Atoms,
	restrictions []ActivationRestriction,
	triggerFrequency *TriggerFrequencyRestriction,
	sourceCostReduction *SourceAbilityCostReductionSyntax,
) {
	if recognizeImpulseExileSequence(sentences) {
		return
	}
	legacyEffects := 0
	currentEffects := 0
	unrecognizedSibling := false
	var riderCandidates []int
	for i := range sentences {
		if sentences[i].StaticRule != nil ||
			sourceCostReduction != nil && sentences[i].Span == sourceCostReduction.Span ||
			spanInsideActivationRestriction(sentences[i].Span, restrictions) ||
			spanInsideTriggerFrequency(sentences[i].Span, triggerFrequency) {
			continue
		}
		tokens := semanticEffectTokens(sentences[i].Tokens)
		count := legacyEffectCount(tokens, atoms)
		legacyEffects += count
		sentences[i].LegacyEffects = count > 0
		sentences[i].Targets = parseTargets(tokens, atoms)
		sentences[i].Effects = parseEffects(sentences[i], tokens, atoms)
		recognizeTargetOpponentHandManaSentence(&sentences[i])
		collapseManaSpendRiderSentence(&sentences[i], tokens)
		currentEffects += len(sentences[i].Effects)
		if len(tokens) > 0 && len(sentences[i].Effects) == 0 &&
			len(atoms.KeywordsWithin(tokens)) == 0 && count == 0 &&
			!effectWordsAt(tokens, 0, "activate", "only", "if") {
			if isRegenerationRiderTokens(tokens) {
				riderCandidates = append(riderCandidates, i)
			} else {
				unrecognizedSibling = true
			}
		}
	}
	recognizeShuffleRevealPermanentSequence(sentences)
	if currentEffects == 1 && unrecognizedSibling {
		for i := range sentences {
			for j := range sentences[i].Effects {
				sentences[i].Effects[j].Exact = false
				sentences[i].Effects[j].HasUnrecognizedSibling = true
			}
		}
	}
	if len(riderCandidates) > 0 {
		creditRegenerationRider(sentences, riderCandidates, unrecognizedSibling)
	}
	if legacyEffects <= 1 {
		return
	}
	for i := range sentences {
		for j := range sentences[i].Effects {
			sentences[i].Effects[j].RequiresOrderedLowering = true
		}
	}
}

// creditRegenerationRider folds one or more "It/They can't be regenerated."
// rider sentences onto the ability's lone destroy effect: it sets
// PreventRegeneration plus a coverage span on the destroy and marks the rider
// sentences so reference and coverage scans credit them. It credits only when
// the ability holds exactly one destroy effect, that destroy is exact, and no
// other sentence is unrecognized; otherwise the riders stay uncredited and the
// card fails closed at the lowering coverage check. Sibling effects other than
// the lone destroy are permitted (for example a "...creates a token" rider that
// destruction spells such as Pongify pair with the regeneration clause), because
// the rider's pronoun subject can only denote the destroyed permanent and the
// rider span is covered independently of those siblings.
func creditRegenerationRider(sentences []Sentence, riderCandidates []int, unrecognizedSibling bool) {
	if unrecognizedSibling {
		return
	}
	destroy := loneDestroyEffect(sentences)
	if destroy == nil || !destroy.Exact {
		return
	}
	riderSpan := sentences[riderCandidates[0]].Span
	for _, index := range riderCandidates[1:] {
		if sentences[index].Span.End.Offset > riderSpan.End.Offset {
			riderSpan.End = sentences[index].Span.End
		}
	}
	destroy.PreventRegeneration = true
	destroy.RegenerationRiderSpan = riderSpan
	for _, index := range riderCandidates {
		sentences[index].RegenerationRider = true
	}
}

// loneDestroyEffect returns the single EffectDestroy across the sentences, or nil
// when the sentences hold zero or more than one destroy effect. Sibling effects
// of other kinds are permitted and ignored; only the count of destroy effects
// constrains the result so a regeneration rider can fold onto a destruction
// clause that is accompanied by a recognized non-destroy effect.
func loneDestroyEffect(sentences []Sentence) *EffectSyntax {
	var found *EffectSyntax
	for i := range sentences {
		for j := range sentences[i].Effects {
			if sentences[i].Effects[j].Kind != EffectDestroy {
				continue
			}
			if found != nil {
				return nil
			}
			found = &sentences[i].Effects[j]
		}
	}
	return found
}

// isRegenerationRiderTokens reports whether the sentence tokens are a
// regeneration rider restricted to the pronoun forms "It can't be regenerated"
// and "They can't be regenerated". Pronoun-only forms avoid introducing phantom
// targets that subject phrases ("that creature", "a creature destroyed this
// way") would otherwise contribute to the compiled target set.
func isRegenerationRiderTokens(tokens []shared.Token) bool {
	if !effectWordsAt(tokens, 0, "it", "can't", "be", "regenerated") &&
		!effectWordsAt(tokens, 0, "they", "can't", "be", "regenerated") {
		return false
	}
	rest := tokens[4:]
	for i := range rest {
		if rest[i].Kind != shared.Period {
			return false
		}
	}
	return true
}

func spanInsideActivationRestriction(span shared.Span, restrictions []ActivationRestriction) bool {
	for i := range restrictions {
		if spanCovers(restrictions[i].Span, span) || spanCovers(span, restrictions[i].Span) {
			return true
		}
	}
	return false
}

func spanInsideTriggerFrequency(span shared.Span, triggerFrequency *TriggerFrequencyRestriction) bool {
	if triggerFrequency == nil {
		return false
	}
	return spanCovers(triggerFrequency.Span, span) || spanCovers(span, triggerFrequency.Span)
}

func semanticEffectTokens(tokens []shared.Token) []shared.Token {
	result := make([]shared.Token, 0, len(tokens))
	depth := 0
	quoted := false
	for _, token := range tokens {
		switch token.Kind {
		case shared.LeftParen:
			if !quoted {
				depth++
			}
		case shared.RightParen:
			if !quoted && depth > 0 {
				depth--
			}
		case shared.Quote:
			quoted = !quoted
		default:
			if depth == 0 && !quoted {
				result = append(result, token)
			}
		}
	}
	return result
}

// tokensBeforeOffset returns the tokens that end at or before the given source
// offset, preserving order. It is used to scope the recipient Selection of a
// trailing-amount damage clause to the tokens before the amount phrase so the
// amount's counted subject does not contaminate the recipient. Returning a
// contiguous prefix (rather than deleting the amount tokens in place) keeps the
// recipient span from bridging across the removed phrase to later punctuation.
func tokensBeforeOffset(tokens []shared.Token, offset int) []shared.Token {
	result := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Span.End.Offset > offset {
			break
		}
		result = append(result, token)
	}
	return result
}

func parseEffects(sentence Sentence, tokens []shared.Token, atoms Atoms) []EffectSyntax {
	if effects, ok := parseLibraryTopReorderEffect(sentence, tokens, atoms); ok {
		return effects
	}
	if effects, ok := parsePlayerProtectionEffects(sentence, tokens, atoms); ok {
		return effects
	}
	if effects, ok := parseGroupPhaseOutEffect(sentence, tokens, atoms); ok {
		return effects
	}
	if effects, ok := parseAdditionalLandPlaysEffect(sentence, tokens, atoms); ok {
		return effects
	}
	indices := effectIndices(tokens, atoms)
	requiresOrderedLowering := legacyEffectCount(tokens, atoms) > 1
	effects := make([]EffectSyntax, 0, len(indices))
	for effectIndex, tokenIndex := range indices {
		clauseEnd := resolvingClauseEnd(tokens, indices, effectIndex)
		ownershipStart := resolvingClauseStart(tokens, indices, effectIndex)
		ownership := tokens[ownershipStart:clauseEnd]
		clause := tokens[tokenIndex+1 : clauseEnd]
		clause, delayed := cutDelayedTiming(clause)
		if delayed == DelayedTimingNone {
			delayed = leadingDelayedTiming(tokens[ownershipStart:tokenIndex])
		}
		power, toughness := parsePTChange(clause)
		counterKind, counterKnown := parseCounterPlacement(clause, atoms)
		span := shared.SpanOf(clause)
		ownershipSpan := shared.SpanOf(ownership)
		toZone := firstZone(atoms, span, ZoneRoleTo)
		if ambiguousZoneChoice(ownership, atoms, span) {
			toZone = zone.None
		}
		staticSubject := parseEffectStaticSubject(ownership, atoms)
		payment := parseEffectPayment(tokens, atoms)
		connection, connectionSpan := effectConnection(tokens, indices, effectIndex)
		optional, optionalSpan := effectOptional(tokens, tokenIndex)
		context := effectContextAt(tokens, tokenIndex, atoms)
		if effectIndex > 0 && !effectHasExplicitSubject(tokens, tokenIndex) &&
			effects[len(effects)-1].Context != EffectContextController {
			context = EffectContextPriorSubject
		}
		durationTokens := ownership
		nextConnection := EffectConnectionNone
		if effectIndex+1 < len(indices) {
			nextConnection, _ = effectConnection(tokens, indices, effectIndex+1)
			if nextConnection == EffectConnectionAnd &&
				durationScopesAcrossAnd(effectKindAt(tokens, tokenIndex), effectKindAt(tokens, indices[effectIndex+1])) {
				durationTokens = tokens
			}
		}
		kind := effectKindAt(tokens, tokenIndex)
		entersColorChoice, entersColorChoiceExclude := entersColorChoiceSyntax(kind, clause)
		tokenPower, tokenToughness, tokenPTKnown := parseTokenPowerToughness(kind, clause)
		amount := parseEffectAmount(kind, clause, atoms)
		if forEach, ok := parseCreateForEachAmount(kind, context, tokenPTKnown, tokens[ownershipStart:tokenIndex], amount, atoms); ok {
			amount = forEach
		}
		// A deal-damage clause whose amount is a trailing "where X is the number
		// of ..." count phrase ("deals X damage to each creature, where X is the
		// number of Gates you control.") embeds the counted-subject selector in
		// the same clause as the recipient group. parseSelection scans the span
		// of its tokens, so leaving the count phrase in would fold the count
		// subject's filters (here "Gate" and "you control") into the recipient,
		// and merely deleting the count tokens would still bridge the span across
		// the trailing sentence period. The recipient group is exactly the run of
		// tokens before the trailing count phrase, so scope the recipient
		// Selection to those, leaving the count subject to the amount's own
		// selector.
		selectionClause := clause
		if kind == EffectDealDamage && amount.DynamicForm == EffectDynamicAmountFormWhereX {
			selectionClause = tokensBeforeOffset(clause, amount.Span.Start.Offset)
		}
		effects = append(effects, EffectSyntax{
			Kind:                     kind,
			Context:                  context,
			Connection:               connection,
			ConnectionSpan:           connectionSpan,
			Span:                     sentence.Span,
			VerbSpan:                 tokens[tokenIndex].Span,
			ClauseSpan:               ownershipSpan,
			Text:                     sentence.Text,
			Tokens:                   append([]shared.Token(nil), ownership...),
			Duration:                 parseEffectDuration(durationTokens, atoms),
			DelayedTiming:            delayed,
			Selection:                parseSelection(selectionClause, atoms),
			DamageRecipientPair:      parseDamageRecipientPair(kind, clause, atoms),
			Amount:                   amount,
			PowerDelta:               power,
			ToughnessDelta:           toughness,
			TokenPower:               tokenPower,
			TokenToughness:           tokenToughness,
			TokenPTKnown:             tokenPTKnown,
			TokenKeywords:            parseTokenKeywords(kind, clause, atoms),
			TokenName:                parseTokenName(kind, clause),
			TokenChoice:              parseTokenChoice(kind, clause),
			StaticSubject:            staticSubject,
			CounterKind:              counterKind,
			CounterKnown:             counterKnown,
			CounterRecipientAttached: counterRecipientAttached(kind, counterKnown, clause),
			FromZone:                 firstZone(atoms, span, ZoneRoleFrom),
			ToZone:                   toZone,
			Destination:              parseEffectDestination(ownership),
			EntersTapped:             effectWordsAtAny(ownership, "battlefield", "tapped"),
			EntersTappedSelf:         entersTappedSelfSyntax(kind, clause),
			EntersColorChoice:        entersColorChoice,
			EntersColorChoiceExclude: entersColorChoiceExclude,
			EntersTypeChoice:         entersTypeChoiceSyntax(kind, clause),
			EntersWithCounters:       entersWithCountersSyntax(kind, clause),
			UnderYourControl:         effectContainsWords(normalizedWords(ownership), "under", "your", "control"),
			CastAsAdventure:          effectContainsWords(normalizedWords(clause), "as", "an", "adventure"),
			Negated:                  effectIsNegated(tokens, tokenIndex),
			Optional:                 optional,
			OptionalSpan:             optionalSpan,
			LifeObject:               gainLoseLifeObject(kind, clause),
			Symbol:                   firstEffectSymbol(clause),
			Mana:                     parseEffectMana(kind, clause, nextConnection != EffectConnectionNone),
			Replacement:              parseEffectReplacement(ownership, atoms),
			References:               referencesInSpan(atoms, ownershipSpan),
			SubjectReferences:        referencesInSpan(atoms, shared.SpanOf(tokens[ownershipStart:tokenIndex])),
			Targets:                  targetsInSpan(sentence.Targets, ownershipSpan),
			SubjectTargets:           targetsInSpan(sentence.Targets, shared.SpanOf(tokens[ownershipStart:tokenIndex])),
			Payment:                  payment,
			RequiresOrderedLowering:  requiresOrderedLowering,
		})
	}

	for i := range effects {
		effects[i].Divided = dividedDamageEffect(&effects[i])
		effects[i].DamageRecipientReference = damageRecipientReference(&effects[i])
		effects[i].SelfDamageRiderValue, effects[i].HasSelfDamageRider = damageSelfRider(&effects[i])
		effects[i].TargetControllerDamageRiderValue, effects[i].TargetControllerDamageRiderRecipient = damageTargetControllerRider(&effects[i])
		effects[i].SecondTargetDamageRiderValue, effects[i].HasSecondTargetDamageRider = damageSecondTargetRider(&effects[i])
		effects[i].Dig = parseDigPut(&effects[i])
		effects[i].HandLibraryPut = parseHandLibraryPut(&effects[i])
		effects[i].HandDiscard = parseHandDiscard(&effects[i])
		effects[i].DiscardEntireHand = parseDiscardEntireHand(&effects[i])
		effects[i].SearchSplit = parseSearchSplitPut(&effects[i])
		effects[i].GraveyardZoneExile = parseGraveyardZoneExile(&effects[i])
		effects[i].Exact = exactEffectSyntax(&effects[i])
		if recognizeTargetOpponentHandMana(&effects[i]) {
			effects[i].Exact = true
		}
		if recognizeControlledCountMana(&effects[i]) {
			effects[i].Exact = true
		}
		effects[i].TokenCopyOfTarget = exactCreateCopyTokenEffectSyntax(&effects[i])
		effects[i].Mana.LegacyBodyExact = legacyExactManaBody(&effects[i], sentence)
		if effects[i].Kind == EffectSearch {
			effects[i].UnsupportedDetail = searchUnsupportedDetail(&effects[i])
			effects[i].SearchSharedSubtype = searchSharedSubtypeRider(&effects[i])
			effects[i].SearchDestination = searchDestinationPosition(&effects[i])
		}
	}
	return effects
}

func parseLibraryTopReorderEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	amount, ok := matchLibraryTopReorder(tokens, atoms)
	if !ok {
		return nil, false
	}
	effect := EffectSyntax{
		Kind:       EffectReorderLibraryTop,
		Context:    EffectContextController,
		Span:       sentence.Span,
		VerbSpan:   tokens[0].Span,
		ClauseSpan: sentence.Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Amount:     amount,
		References: referencesInSpan(atoms, sentence.Span),
	}
	effect.Exact = exactEffectSyntax(&effect)
	return []EffectSyntax{effect}, true
}

func matchLibraryTopReorder(tokens []shared.Token, atoms Atoms) (EffectAmountSyntax, bool) {
	if len(tokens) != 18 ||
		!effectWordsAt(tokens, 0, "look", "at", "the", "top") ||
		!effectWordsAt(tokens, 5, "cards", "of", "your", "library") ||
		tokens[9].Kind != shared.Comma ||
		!effectWordsAt(tokens, 10, "then", "put", "them", "back", "in", "any", "order") ||
		tokens[17].Kind != shared.Period {
		return EffectAmountSyntax{}, false
	}
	amount := parseEffectAmount(EffectReorderLibraryTop, tokens[4:5], atoms)
	return amount, amount.Known && amount.Value > 0
}

func recognizeImpulseExileSequence(sentences []Sentence) bool {
	if len(sentences) != 2 ||
		!strings.EqualFold(strings.TrimSpace(sentences[0].Text), "Exile the top three cards of your library.") ||
		!strings.EqualFold(strings.TrimSpace(sentences[1].Text), "You may play them this turn.") {
		return false
	}
	span := shared.Span{Start: sentences[0].Span.Start, End: sentences[1].Span.End}
	sentences[0].Effects = []EffectSyntax{{
		Kind:       EffectImpulseExile,
		Context:    EffectContextController,
		Span:       span,
		ClauseSpan: span,
		Text:       sentences[0].Text + " " + sentences[1].Text,
		Tokens:     append(append([]shared.Token(nil), sentences[0].Tokens...), sentences[1].Tokens...),
		Amount:     EffectAmountSyntax{Value: 3, Known: true},
		Duration:   EffectDurationThisTurn,
		Exact:      true,
	}}
	return true
}

func recognizeTargetOpponentHandMana(effect *EffectSyntax) bool {
	if effect.Kind != EffectAddMana ||
		!strings.EqualFold(strings.TrimSpace(exactEffectClauseText(effect)), "Add {R} for each card in target opponent's hand.") {
		return false
	}
	effect.Amount = EffectAmountSyntax{
		DynamicKind: EffectDynamicAmountCount,
		DynamicForm: EffectDynamicAmountFormForEach,
		Multiplier:  1,
		Selection: &SelectionSyntax{
			Kind:       SelectionCard,
			Controller: SelectionControllerOpponent,
			Zone:       zone.Hand,
		},
	}
	effect.Mana = EffectManaSyntax{
		Symbols:     []string{"{R}"},
		Colors:      []mana.Color{mana.R},
		ColorsKnown: true,
	}
	return true
}

func recognizeTargetOpponentHandManaSentence(sentence *Sentence) {
	if len(sentence.Effects) != 1 ||
		!recognizeTargetOpponentHandMana(&sentence.Effects[0]) ||
		len(sentence.Targets) != 1 {
		return
	}
	target := sentence.Targets[0]
	target.Cardinality = TargetCardinalitySyntax{Min: 1, Max: 1}
	target.Selection = SelectionSyntax{Kind: SelectionOpponent}
	target.Exact = true
	sentence.Targets[0] = target
	sentence.Effects[0].Targets = []TargetSyntax{target}
}

// recognizeControlledCountMana types an "Add <mana> for each <permanent> you
// control" add-mana body (Cabal Coffers, Gaea's Cradle, Serra's Sanctum) whose
// produced amount scales with a battlefield permanent count. The "for each
// <permanent>" suffix is already typed onto effect.Amount as a dynamic
// battlefield count by parseEffectAmount; the leading mana symbol, however, is
// left unparsed because parseEffectMana rejects the trailing count clause. This
// re-parses just the symbols before the count phrase and records the produced
// color, so the lowerer can emit one mana per counted permanent. It fires only
// for a single fixed produced color over a battlefield (non-zone) count, so
// choice, any-color, and multi-symbol outputs stay fail-closed.
func recognizeControlledCountMana(effect *EffectSyntax) bool {
	if effect.Kind != EffectAddMana ||
		effect.Amount.DynamicKind != EffectDynamicAmountCount ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormForEach ||
		effect.Amount.Multiplier < 1 ||
		effect.Amount.Selection == nil ||
		effect.Amount.Selection.Zone != zone.None {
		return false
	}
	body := manaBodyBeforeAmount(effect)
	if len(body) == 0 {
		return false
	}
	parsed := parseEffectMana(EffectAddMana, body, true)
	if !parsed.ColorsKnown || len(parsed.Colors) != 1 || parsed.Choice {
		return false
	}
	effect.Mana = parsed
	return true
}

// manaBodyBeforeAmount returns the effect tokens that sit after the add-mana
// verb and before the trailing dynamic-count phrase (the produced mana symbols).
func manaBodyBeforeAmount(effect *EffectSyntax) []shared.Token {
	verbEnd := effect.VerbSpan.End.Offset
	amountStart := effect.Amount.Span.Start.Offset
	var body []shared.Token
	for _, token := range effect.Tokens {
		if token.Span.Start.Offset >= verbEnd && token.Span.End.Offset <= amountStart {
			body = append(body, token)
		}
	}
	return body
}

func parseHandDiscard(effect *EffectSyntax) HandDiscardSyntax {
	if effect.Kind != EffectDiscard ||
		effect.Context != EffectContextController ||
		!effect.Amount.Known || effect.Amount.Value < 1 ||
		len(effect.Targets) != 0 ||
		len(effect.References) != 0 ||
		!exactCardCountEffectSyntax(effect, "Discard", "discards", false) {
		return HandDiscardSyntax{}
	}
	return HandDiscardSyntax{Present: true}
}

// parseDiscardEntireHand recognizes a "discard their hand" clause whose affected
// player discards every card in hand. It accepts the controller ("Discard your
// hand"), each-player, each-opponent, and single-target-player forms; the
// amount is unknown because it depends on the player's hand at resolution.
func parseDiscardEntireHand(effect *EffectSyntax) bool {
	if effect.Kind != EffectDiscard ||
		effect.Amount.Known ||
		effect.Negated ||
		effect.Optional {
		return false
	}
	text := strings.TrimSpace(exactEffectClauseText(effect))
	switch effect.Context {
	case EffectContextController:
		return len(effect.Targets) == 0 &&
			(strings.EqualFold(text, "Discard your hand.") ||
				strings.EqualFold(text, "You discard your hand."))
	case EffectContextEachPlayer:
		return len(effect.Targets) == 0 &&
			strings.EqualFold(text, "Each player discards their hand.")
	case EffectContextEachOpponent:
		return len(effect.Targets) == 0 &&
			strings.EqualFold(text, "Each opponent discards their hand.")
	case EffectContextTarget:
		return len(effect.Targets) == 1 && effect.Targets[0].Exact &&
			exactCardCountTargetPlayer(effect.Targets[0].Selection) &&
			strings.EqualFold(text, titleFirstEffectText(effect.Targets[0].Text)+" discards their hand.")
	default:
		return false
	}
}

func parsePlayerProtectionEffects(sentence Sentence, tokens []shared.Token, _ Atoms) ([]EffectSyntax, bool) {
	if strings.TrimSpace(sentence.Text) != "Until your next turn, your life total can't change and you gain protection from everything." {
		return nil, false
	}
	changeIndex, andIndex, gainIndex := -1, -1, -1
	for i := range tokens {
		switch {
		case strings.EqualFold(tokens[i].Text, "change"):
			changeIndex = i
		case changeIndex >= 0 && andIndex < 0 && strings.EqualFold(tokens[i].Text, "and"):
			andIndex = i
		case strings.EqualFold(tokens[i].Text, "gain"):
			gainIndex = i
		default:
		}
	}
	if changeIndex < 0 || andIndex < 0 || gainIndex < 0 || andIndex+1 >= len(tokens) {
		return nil, false
	}
	base := EffectSyntax{
		Span:                    sentence.Span,
		Text:                    sentence.Text,
		Tokens:                  append([]shared.Token(nil), tokens...),
		Duration:                EffectDurationUntilYourNextTurn,
		Context:                 EffectContextController,
		Exact:                   true,
		RequiresOrderedLowering: true,
	}
	life := base
	life.Kind = EffectLifeTotalCantChange
	life.ClauseSpan = shared.Span{Start: sentence.Span.Start, End: tokens[changeIndex].Span.End}
	life.VerbSpan = tokens[changeIndex].Span
	protection := base
	protection.Kind = EffectProtectionFromEverything
	protection.Connection = EffectConnectionAnd
	protection.ConnectionSpan = tokens[andIndex].Span
	protection.ClauseSpan = shared.Span{Start: tokens[andIndex+1].Span.Start, End: sentence.Span.End}
	protection.VerbSpan = tokens[gainIndex].Span
	return []EffectSyntax{life, protection}, true
}

func parseGroupPhaseOutEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	if strings.TrimSpace(sentence.Text) != "All permanents you control phase out." {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:       EffectPhaseOut,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Selection:  parseSelection(tokens, atoms),
		Exact:      true,
	}}, true
}

// parseAdditionalLandPlaysEffect recognizes the controller-scoped grant of one
// or more extra land plays for the turn: "Play an additional land this turn.",
// "You may play an additional land this turn.", and the multi-land "... two
// additional lands ..." / "... up to N additional lands ..." variants. The "you
// may" permission is folded into an unconditional allowance (the player is never
// forced to play the extra land). The static "on each of your turns" form is a
// separate static ability and is not matched here.
func parseAdditionalLandPlaysEffect(sentence Sentence, tokens []shared.Token, _ Atoms) ([]EffectSyntax, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	start := 0
	if len(words) >= 2 && equalWord(words[0], "you") && equalWord(words[1], "may") {
		start = 2
	}
	rest := words[start:]
	// Shortest match: "play an additional land this turn" (6 words).
	if len(rest) < 6 || !equalWord(rest[0], "play") {
		return nil, false
	}
	playToken := rest[0]
	rest = rest[1:]
	if equalWord(rest[0], "up") && len(rest) >= 2 && equalWord(rest[1], "to") {
		rest = rest[2:]
	}
	if len(rest) < 5 {
		return nil, false
	}
	count, ok := additionalLandCountWord(rest[0])
	if !ok || !equalWord(rest[1], "additional") {
		return nil, false
	}
	plural := count != 1
	landWord := "land"
	if plural {
		landWord = "lands"
	}
	if !equalWord(rest[2], landWord) ||
		!equalWord(rest[3], "this") ||
		!equalWord(rest[4], "turn") ||
		len(rest) != 5 {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:       EffectAdditionalLandPlays,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		VerbSpan:   playToken.Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Context:    EffectContextController,
		Duration:   EffectDurationThisTurn,
		Amount:     EffectAmountSyntax{Value: count, Known: true},
		Exact:      true,
	}}, true
}

// additionalLandCountWord reads the extra-land count from the determiner or
// number word preceding "additional land(s)": "a"/"an"/"one" mean a single extra
// land, and small cardinal words ("two", "three", ...) and integer literals give
// their value.
func additionalLandCountWord(token shared.Token) (int, bool) {
	if token.Kind == shared.Integer {
		value, err := strconv.Atoi(token.Text)
		if err != nil || value < 1 {
			return 0, false
		}
		return value, true
	}
	switch strings.ToLower(token.Text) {
	case "a", "an", "one":
		return 1, true
	default:
		return CardinalWordValue(token.Text)
	}
}

func parseHandLibraryPut(effect *EffectSyntax) HandLibraryPutSyntax {
	if effect.Kind != EffectPut ||
		effect.Context != EffectContextController ||
		!effect.Amount.Known || effect.Amount.Value < 1 ||
		effect.FromZone != zone.Hand ||
		effect.ToZone != zone.Library ||
		effect.Destination != EffectDestinationTop ||
		len(effect.Targets) != 0 ||
		!effectContainsWords(normalizedWords(effect.Tokens), "in", "any", "order") {
		return HandLibraryPutSyntax{}
	}
	return HandLibraryPutSyntax{Present: true}
}

// parseDigPut recognizes the impulse put clause "Put N <of them|of those cards>
// into your hand and the <rest|other> into your graveyard." that follows an
// EffectDig look sentence, returning its structured fields. It returns the zero
// DigSyntax for every other effect, including the library-bottom remainder forms
// (which carry an unmodeled ordering rider) so they fail closed. The structured
// fields it sets are revalidated byte-for-byte by exactDigPutEffectSyntax, so an
// over-broad match simply fails the exactness gate.
func parseDigPut(effect *EffectSyntax) DigSyntax {
	if effect.Kind != EffectPut {
		return DigSyntax{}
	}
	verb := slices.IndexFunc(effect.Tokens, func(token shared.Token) bool {
		return token.Span == effect.VerbSpan
	})
	if verb < 0 {
		return DigSyntax{}
	}
	clause := effect.Tokens[verb+1:]
	if len(clause) == 0 ||
		(!equalWord(clause[0], "one") && !equalWord(clause[0], "two") && !equalWord(clause[0], "three")) {
		return DigSyntax{}
	}
	i := 1
	var dig DigSyntax
	switch {
	case effectWordsAt(clause, i, "of", "them"):
		dig.Source = DigSourceThem
		i += 2
	case effectWordsAt(clause, i, "of", "those", "cards"):
		dig.Source = DigSourceThoseCards
		i += 3
	default:
		dig.Source = DigSourceNone
	}
	if !effectWordsAt(clause, i, "into", "your", "hand", "and", "the") {
		return DigSyntax{}
	}
	i += 5
	switch {
	case effectWordsAt(clause, i, "other"):
		dig.Singular = true
		i++
	case effectWordsAt(clause, i, "rest"):
		i++
	default:
		return DigSyntax{}
	}
	if !effectWordsAt(clause, i, "into", "your", "graveyard") {
		return DigSyntax{}
	}
	i += 3
	if i < len(clause) && clause[i].Kind == shared.Period {
		i++
	}
	if i != len(clause) {
		return DigSyntax{}
	}
	dig.Put = true
	return dig
}

// parseSearchSplitPut recognizes the split-destination put clause "put one
// <slot> and the other <slot>" that distributes the cards found by a preceding
// "up to two" library search across two single-card destination slots ("put one
// onto the battlefield tapped and the other into your hand"). It returns the
// zero SearchSplitSyntax for every other effect, including ordinary
// single-destination puts. Each slot is a hand or battlefield (optionally
// tapped) destination; any other wording fails closed. The structured fields it
// sets are revalidated byte-for-byte by the search exactness gate, so an
// over-broad match simply fails recognition.
func parseSearchSplitPut(effect *EffectSyntax) SearchSplitSyntax {
	if effect.Kind != EffectPut {
		return SearchSplitSyntax{}
	}
	verb := slices.IndexFunc(effect.Tokens, func(token shared.Token) bool {
		return token.Span == effect.VerbSpan
	})
	if verb < 0 {
		return SearchSplitSyntax{}
	}
	clause := effect.Tokens[verb+1:]
	if len(clause) == 0 || !equalWord(clause[0], "one") {
		return SearchSplitSyntax{}
	}
	first, i, ok := parseSearchSplitSlot(clause, 1)
	if !ok || !effectWordsAt(clause, i, "and", "the", "other") {
		return SearchSplitSyntax{}
	}
	i += 3
	second, i, ok := parseSearchSplitSlot(clause, i)
	if !ok {
		return SearchSplitSyntax{}
	}
	if i < len(clause) && clause[i].Kind == shared.Period {
		i++
	}
	if i != len(clause) {
		return SearchSplitSyntax{}
	}
	return SearchSplitSyntax{Present: true, First: first, Second: second}
}

// parseSearchSplitSlot reads one destination slot of a split put clause starting
// at index i: "onto the battlefield" with an optional trailing "tapped", or
// "into your hand". It returns the slot, the index just past it, and whether a
// slot was recognized.
func parseSearchSplitSlot(clause []shared.Token, i int) (SearchSplitSlot, int, bool) {
	switch {
	case effectWordsAt(clause, i, "onto", "the", "battlefield"):
		slot := SearchSplitSlot{ToZone: zone.Battlefield}
		i += 3
		if effectWordsAt(clause, i, "tapped") {
			slot.EntersTapped = true
			i++
		}
		return slot, i, true
	case effectWordsAt(clause, i, "into", "your", "hand"):
		return SearchSplitSlot{ToZone: zone.Hand}, i + 3, true
	default:
		return SearchSplitSlot{}, i, false
	}
}

// parseDamageRecipientPair recognizes the dual-recipient group-damage wording
// "deals N damage to each X and each Y" and returns the two recipient groups as
// separate selections. It returns nil for every other effect so the recipient
// stays single. The recipient is identified as the tokens after "damage to";
// it must split into exactly two "each <group>" phrases joined by a single
// top-level "and", and each phrase is parsed by the same parseSelection used for
// a lone group recipient. The downstream exactness gate reconstructs both halves
// and compares them byte-for-byte, so an over-broad split simply fails closed.
func parseDamageRecipientPair(kind EffectKind, clause []shared.Token, atoms Atoms) []SelectionSyntax {
	if kind != EffectDealDamage {
		return nil
	}
	recipient, ok := damageRecipientTokens(clause)
	if !ok {
		return nil
	}
	left, right, ok := splitEachAndEach(recipient)
	if !ok {
		return nil
	}
	return []SelectionSyntax{
		parseSelection(left, atoms),
		parseSelection(right, atoms),
	}
}

// damageRecipientTokens returns the recipient tokens of a deal-damage clause:
// everything after the first "damage to", with a trailing period removed. It
// fails closed when "damage" is not immediately followed by "to" (for example
// the dynamic "damage equal to ... to ..." form), leaving such wordings to other
// paths.
func damageRecipientTokens(clause []shared.Token) ([]shared.Token, bool) {
	for i := 0; i+1 < len(clause); i++ {
		if equalWord(clause[i], "damage") && equalWord(clause[i+1], "to") {
			recipient := clause[i+2:]
			if len(recipient) > 0 && recipient[len(recipient)-1].Kind == shared.Period {
				recipient = recipient[:len(recipient)-1]
			}
			if len(recipient) == 0 {
				return nil, false
			}
			return recipient, true
		}
	}
	return nil, false
}

// damageRecipientReference recognizes a damage recipient that is the controller
// or owner of a referenced object (the prior removal target): "deals N damage to
// its controller", "... to its owner", "... to that <object>'s controller", or
// "... to that <object>'s owner". It uses the effect's own Tokens (the clause
// span) so the recipient is read from the verb clause alone. It returns None for
// every other recipient (a target, a group, or a dual recipient), leaving those
// to their existing paths.
func damageRecipientReference(effect *EffectSyntax) DamageRecipientReferenceKind {
	if effect.Kind != EffectDealDamage {
		return DamageRecipientReferenceNone
	}
	recipient, ok := damageRecipientTokens(effect.Tokens)
	if !ok {
		return DamageRecipientReferenceNone
	}
	// "deals N damage to you" names the source's own controller. The lone "you"
	// recipient carries no object subject, so it is recognized before the
	// referenced-object controller/owner forms below.
	if len(recipient) == 1 && equalWord(recipient[0], "you") {
		return DamageRecipientReferenceYou
	}
	if len(recipient) < 2 {
		return DamageRecipientReferenceNone
	}
	role, ok := referencedControllerOwnerRecipient(recipient)
	if !ok {
		return DamageRecipientReferenceNone
	}
	return role
}

// damageSelfRider recognizes a "... and N damage to you" self-damage rider
// appended to a deal-damage clause whose primary recipient is its single target,
// as in "deals 4 damage to any target and 2 damage to you." It returns the fixed
// rider amount N (>= 1) and ok=true only when the clause ends with the exact
// "and <number> damage to you" suffix. It fails closed for every other ending
// (a non-"you" recipient, a missing leading "and", a non-numeric amount), so the
// dual-group "each X and each Y" recipient and the standalone "to you" recipient
// keep their existing paths.
func damageSelfRider(effect *EffectSyntax) (int, bool) {
	if effect.Kind != EffectDealDamage {
		return 0, false
	}
	tokens := effect.Tokens
	if len(tokens) > 0 && tokens[len(tokens)-1].Kind == shared.Period {
		tokens = tokens[:len(tokens)-1]
	}
	n := len(tokens)
	if n < 5 {
		return 0, false
	}
	if !equalWord(tokens[n-1], "you") ||
		!equalWord(tokens[n-2], "to") ||
		!equalWord(tokens[n-3], "damage") ||
		!equalWord(tokens[n-5], "and") {
		return 0, false
	}
	value, ok := damageRiderAmountValue(tokens[n-4])
	if !ok || value < 1 {
		return 0, false
	}
	return value, true
}

// damageRiderAmountValue reads the fixed numeric value of a self-damage rider
// amount token, accepting both an integer literal ("2") and a small cardinal
// word ("two"). It returns ok=false for any non-numeric token.
func damageRiderAmountValue(token shared.Token) (int, bool) {
	if token.Kind == shared.Integer {
		value, err := strconv.Atoi(token.Text)
		if err != nil {
			return 0, false
		}
		return value, true
	}
	return CardinalWordValue(token.Text)
}

// damageTargetControllerRider recognizes a "... and B damage to that creature's
// controller/owner" rider appended to a single-target deal-damage clause, as in
// "Chandra's Outrage deals 4 damage to target creature and 2 damage to that
// creature's controller." It returns the fixed rider amount B (>= 1) and the
// recipient role (controller or owner of the primary target). It fails closed
// (None) for every other ending, including the "to you" self rider and the
// dual-group "each X and each Y" recipient, which keep their existing paths.
func damageTargetControllerRider(effect *EffectSyntax) (int, DamageRecipientReferenceKind) {
	value, recipient, _ := targetControllerDamageRiderTokens(effect)
	return value, recipient
}

// targetControllerDamageRiderTokens detects the "... and B damage to that
// creature's controller/owner" rider suffix and returns the rider amount, the
// recipient role, and the recipient tokens (for exact reconstruction). It fails
// closed (ok=false) for every other ending.
func targetControllerDamageRiderTokens(effect *EffectSyntax) (int, DamageRecipientReferenceKind, []shared.Token) {
	if effect.Kind != EffectDealDamage {
		return 0, DamageRecipientReferenceNone, nil
	}
	tokens := effect.Tokens
	if len(tokens) > 0 && tokens[len(tokens)-1].Kind == shared.Period {
		tokens = tokens[:len(tokens)-1]
	}
	n := len(tokens)
	// The recipient phrase is "its controller/owner" (2 tokens) or "that
	// <noun>'s controller/owner" (3 tokens), preceded by "and <number> damage
	// to" (4 tokens).
	for _, recipientLen := range []int{2, 3} {
		if n < recipientLen+4 {
			continue
		}
		recipient := tokens[n-recipientLen:]
		role, ok := referencedControllerOwnerRecipient(recipient)
		if !ok {
			continue
		}
		head := n - recipientLen
		if !equalWord(tokens[head-4], "and") ||
			!equalWord(tokens[head-2], "damage") ||
			!equalWord(tokens[head-1], "to") {
			continue
		}
		value, ok := damageRiderAmountValue(tokens[head-3])
		if !ok || value < 1 {
			continue
		}
		return value, role, recipient
	}
	return 0, DamageRecipientReferenceNone, nil
}

// referencedControllerOwnerRecipient reports whether the recipient tokens name
// the controller or owner of a referenced object — "its controller", "its
// owner", "that <noun>'s controller", or "that <noun>'s owner" — and returns
// the matching recipient role. It fails closed (None) for any other phrase.
func referencedControllerOwnerRecipient(recipient []shared.Token) (DamageRecipientReferenceKind, bool) {
	if len(recipient) < 2 {
		return DamageRecipientReferenceNone, false
	}
	role := recipient[len(recipient)-1]
	subject := recipient[:len(recipient)-1]
	subjectIsReferencedObject := len(subject) == 1 && equalWord(subject[0], "its") ||
		len(subject) == 2 && equalWord(subject[0], "that") && referencePossessiveObjectNoun(subject[1])
	if !subjectIsReferencedObject {
		return DamageRecipientReferenceNone, false
	}
	switch {
	case equalWord(role, "controller"):
		return DamageRecipientReferenceController, true
	case equalWord(role, "owner"):
		return DamageRecipientReferenceOwner, true
	default:
		return DamageRecipientReferenceNone, false
	}
}

// damageSecondTargetRider recognizes a "... and B damage to <second target>"
// rider appended to a single-target deal-damage clause whose second clause names
// its own target, as in "Hungry Flames deals 3 damage to target creature and 2
// damage to target player or planeswalker." It requires the clause to carry
// exactly two parsed targets and the rider suffix "and <number> damage to" to
// land immediately before the second target's span. It returns the fixed rider
// amount B (>= 1) and ok=true, failing closed for every other shape so single-
// target and group-recipient clauses keep their existing paths.
func damageSecondTargetRider(effect *EffectSyntax) (int, bool) {
	if effect.Kind != EffectDealDamage || len(effect.Targets) != 2 {
		return 0, false
	}
	tokens := effect.Tokens
	if len(tokens) > 0 && tokens[len(tokens)-1].Kind == shared.Period {
		tokens = tokens[:len(tokens)-1]
	}
	secondStart := effect.Targets[1].Span.Start.Offset
	for i := 0; i+4 < len(tokens); i++ {
		if !equalWord(tokens[i], "and") {
			continue
		}
		value, ok := damageRiderAmountValue(tokens[i+1])
		if !ok || value < 1 {
			continue
		}
		if !equalWord(tokens[i+2], "damage") || !equalWord(tokens[i+3], "to") {
			continue
		}
		if tokens[i+4].Span.Start.Offset == secondStart {
			return value, true
		}
	}
	return 0, false
}

// splitEachAndEach splits recipient tokens at a single top-level "and" into two
// phrases that each begin with "each". It fails closed for any other shape (no
// "and", more than one "and", or a half that does not start with "each"), so
// single recipients and unsupported compounds are left to the single-recipient
// path.
func splitEachAndEach(recipient []shared.Token) (left, right []shared.Token, ok bool) {
	andIndex := -1
	for i := range recipient {
		if equalWord(recipient[i], "and") {
			if andIndex != -1 {
				return nil, nil, false
			}
			andIndex = i
		}
	}
	if andIndex <= 0 || andIndex >= len(recipient)-1 {
		return nil, nil, false
	}
	left = recipient[:andIndex]
	right = recipient[andIndex+1:]
	if !equalWord(left[0], "each") || !equalWord(right[0], "each") {
		return nil, nil, false
	}
	return left, right, true
}

func legacyExactManaBody(effect *EffectSyntax, sentence Sentence) bool {
	if effect.Kind != EffectAddMana || len(semanticEffectTokens(sentence.Tokens)) != len(sentence.Tokens) {
		return false
	}
	direct := len(effect.Tokens) > 0 && equalWord(effect.Tokens[0], "add")
	optionalController := len(effect.Tokens) > 2 &&
		effectWordsAt(effect.Tokens, 0, "you", "may", "add")
	if !direct && !optionalController {
		return false
	}
	return effect.Mana.AnyColor || effect.Mana.CommanderIdentity || effect.Mana.LandsProduce || effect.Mana.FilterPair || len(effect.Mana.Symbols) != 0
}

func legacyEffectCount(tokens []shared.Token, atoms Atoms) int {
	count := 0
	for i := range tokens {
		if legacyEffectKindAt(tokens, i) != EffectUnknown &&
			!atoms.SelfNameAt(tokens[i].Span) &&
			!effectWithinCondition(tokens, i) {
			count++
		}
	}
	return count
}

func effectWithinCondition(tokens []shared.Token, index int) bool {
	for i := index - 1; i >= 0; i-- {
		if tokens[i].Kind == shared.Comma || tokens[i].Kind == shared.Period || tokens[i].Kind == shared.Semicolon {
			return false
		}
		if equalWord(tokens[i], "if") || equalWord(tokens[i], "unless") {
			return true
		}
	}
	return false
}

func legacyEffectKindAt(tokens []shared.Token, index int) EffectKind {
	if equalWord(tokens[index], "look") {
		return EffectManifestDread
	}
	if cantBeBlockedThisTurnVerbAt(tokens, index) {
		return EffectCantBeBlocked
	}
	kind := effectWordKind(tokens[index])
	switch {
	case kind == EffectGrantKeyword && index >= 2 &&
		(equalWord(tokens[index-2], "opponent") || equalWord(tokens[index-2], "opponents")) &&
		equalWord(tokens[index-1], "you"):
		return EffectUnknown
	case kind == EffectEnterTapped && index+1 < len(tokens) && equalWord(tokens[index+1], "prepared"):
		return EffectEnterPrepared
	case kind == EffectCast && index > 0 && (equalWord(tokens[index-1], "was") || equalWord(tokens[index-1], "were")):
		return EffectUnknown
	case kind == EffectCounter && !counterVerbAt(tokens, index):
		return EffectUnknown
	case kind == EffectGain && index+1 < len(tokens) && equalWord(tokens[index+1], "control"):
		return EffectGainControl
	case kind == EffectDouble && index+1 < len(tokens) && equalWord(tokens[index+1], "strike"):
		return EffectUnknown
	case kind == EffectGrantKeyword && priorPTChange(tokens, index):
		return EffectUnknown
	default:
		return kind
	}
}

// entersColorChoiceSyntax recognizes the self entry color-choice clause "choose
// a color ." (unconstrained) or "choose a color other than <color> ." (a single
// forbidden basic color, the Gate/Thriving land cycle) following an "As this
// <permanent> enters," verb. The enters verb is shared by many entry constructs,
// so this matches only these exact color-choice clauses; non-color choices fail
// closed. The returned color is the forbidden color for the "other than" form,
// or empty otherwise.
func entersColorChoiceSyntax(kind EffectKind, clause []shared.Token) (bool, mana.Color) {
	if kind != EffectEnterTapped {
		return false, ""
	}
	body := clause
	if len(body) > 0 && body[0].Kind == shared.Comma {
		body = body[1:]
	}
	if len(body) == 4 &&
		equalWord(body[0], "choose") &&
		equalWord(body[1], "a") &&
		equalWord(body[2], "color") &&
		body[3].Text == "." {
		return true, ""
	}
	if len(body) == 7 &&
		equalWord(body[0], "choose") &&
		equalWord(body[1], "a") &&
		equalWord(body[2], "color") &&
		equalWord(body[3], "other") &&
		equalWord(body[4], "than") &&
		body[6].Text == "." {
		if forbidden, ok := basicColorWord(body[5]); ok {
			return true, forbidden
		}
	}
	return false, ""
}

// entersTypeChoiceSyntax recognizes the self entry creature-type-choice clause
// "choose a creature type ." following an "As this <permanent> enters," verb.
// The enters verb is shared by many entry constructs, so this matches only this
// exact clause; other choices fail closed.
func entersTypeChoiceSyntax(kind EffectKind, clause []shared.Token) bool {
	if kind != EffectEnterTapped {
		return false
	}
	body := clause
	if len(body) > 0 && body[0].Kind == shared.Comma {
		body = body[1:]
	}
	return len(body) == 5 &&
		equalWord(body[0], "choose") &&
		equalWord(body[1], "a") &&
		equalWord(body[2], "creature") &&
		equalWord(body[3], "type") &&
		body[4].Text == "."
}

// basicColorWord maps a single English basic color word to its typed mana color.
// It fails closed on any other token so unrecognized color words leave the entry
// choice unconstrained.
func basicColorWord(token shared.Token) (mana.Color, bool) {
	switch {
	case equalWord(token, "white"):
		return mana.W, true
	case equalWord(token, "blue"):
		return mana.U, true
	case equalWord(token, "black"):
		return mana.B, true
	case equalWord(token, "red"):
		return mana.R, true
	case equalWord(token, "green"):
		return mana.G, true
	default:
		return "", false
	}
}

func entersWithCountersSyntax(kind EffectKind, clause []shared.Token) bool {
	if kind != EffectEnterTapped || len(clause) < 4 ||
		!equalWord(clause[0], "with") ||
		!equalWord(clause[len(clause)-3], "on") ||
		!equalWord(clause[len(clause)-2], "it") ||
		clause[len(clause)-1].Text != "." {
		return false
	}
	for _, token := range clause[1 : len(clause)-3] {
		if equalWord(token, "counter") || equalWord(token, "counters") {
			return true
		}
	}
	return false
}

// entersTappedSelfSyntax recognizes a self enters-tapped instruction such as
// "This land enters tapped." or "Nyx Lotus enters tapped." The enters verb is
// shared by many entry constructs ("As ~ enters, choose ...", "enters with
// counters", "enters tapped and attacking"), so the qualifier following the
// verb must be exactly "tapped" (optionally "the battlefield tapped") to avoid
// classifying unrelated entry effects as a plain tapped entry.
func entersTappedSelfSyntax(kind EffectKind, clause []shared.Token) bool {
	if kind != EffectEnterTapped {
		return false
	}
	body := clause
	if len(body) >= 2 && equalWord(body[0], "the") && equalWord(body[1], "battlefield") {
		body = body[2:]
	}
	return len(body) == 2 && equalWord(body[0], "tapped") && body[1].Text == "."
}

// counterRecipientAttached reports that a counter-placement effect ("put ...
// counter(s) on enchanted creature") targets the permanent the source Aura is
// attached to. It gates on the counter verb and a known counter kind and matches
// only the bare "on enchanted creature" recipient; exact canonical
// reconstruction independently confirms the full clause wording, so any
// additional qualifier leaves the effect inexact and fails closed in lowering.
func counterRecipientAttached(kind EffectKind, counterKnown bool, clause []shared.Token) bool {
	if kind != EffectPut || !counterKnown {
		return false
	}
	return effectHasTokenWords(clause, "on", "enchanted", "creature")
}
