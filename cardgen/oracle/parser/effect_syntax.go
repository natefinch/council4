package parser

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
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
		recognizeOptionalManaPaymentBenefitSequence(&abilities[i])
		recognizeEventPlayerOptionalPaymentSequence(&abilities[i])
		recognizeControllerMandatoryPaymentSequence(&abilities[i])
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
	if recognizeRepeatProcessSequence(sentences, atoms) {
		return
	}
	legacyEffects := 0
	currentEffects := 0
	unrecognizedSibling := false
	var riderCandidates []int
	var chooseColorCandidates []int
	var enchantmentReturnCandidates []int
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
			switch {
			case isRegenerationRiderTokens(tokens) || isThisWayRegenerationRiderTokens(tokens):
				riderCandidates = append(riderCandidates, i)
			case isChosenColorChooseTokens(tokens):
				chooseColorCandidates = append(chooseColorCandidates, i)
			case isEnchantmentReturnRiderTokens(tokens):
				enchantmentReturnCandidates = append(enchantmentReturnCandidates, i)
			default:
				unrecognizedSibling = true
			}
		}
	}
	recognizeShuffleRevealPermanentSequence(sentences)
	if len(chooseColorCandidates) > 0 && !creditChosenColorChoice(sentences, chooseColorCandidates) {
		unrecognizedSibling = true
	}
	if foldedLegacy, foldedEffects, ok := creditTokenCopyGrantRider(sentences, atoms); ok {
		legacyEffects -= foldedLegacy
		currentEffects -= foldedEffects
	}
	if foldedLegacy, foldedEffects, ok := creditCopyChooseNewTargetsRider(sentences, atoms); ok {
		legacyEffects -= foldedLegacy
		currentEffects -= foldedEffects
	}
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
	if len(enchantmentReturnCandidates) > 0 {
		creditEnchantmentReturnRider(sentences, enchantmentReturnCandidates, unrecognizedSibling)
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

// isEnchantmentReturnRiderTokens reports whether the sentence tokens are the
// "It's an enchantment." rider of the Enduring enchantment-creature cycle. The
// parenthetical "(It's not a creature.)" reminder is stripped before sentence
// parsing, so only the bare declaration remains. The rider folds onto a
// preceding return-to-battlefield effect, recording that the returned permanent
// enters as an Enchantment (losing its creature type).
func isEnchantmentReturnRiderTokens(tokens []shared.Token) bool {
	if !effectWordsAt(tokens, 0, "it's", "an", "enchantment") {
		return false
	}
	rest := tokens[3:]
	for i := range rest {
		if rest[i].Kind != shared.Period {
			return false
		}
	}
	return true
}

// creditEnchantmentReturnRider folds one or more "It's an enchantment." rider
// sentences onto the ability's lone return-to-battlefield effect: it sets
// ReturnAsEnchantment plus a coverage span on the return and marks the rider
// sentences so reference and coverage scans credit them. It credits only when
// the ability holds exactly one return-to-battlefield effect and no other
// sentence is unrecognized; otherwise the riders stay uncredited and the card
// fails closed at the lowering coverage check.
func creditEnchantmentReturnRider(sentences []Sentence, riderCandidates []int, unrecognizedSibling bool) {
	if unrecognizedSibling {
		return
	}
	ret := loneReturnToBattlefieldEffect(sentences)
	if ret == nil {
		return
	}
	riderSpan := sentences[riderCandidates[0]].Span
	for _, index := range riderCandidates[1:] {
		if sentences[index].Span.End.Offset > riderSpan.End.Offset {
			riderSpan.End = sentences[index].Span.End
		}
	}
	ret.ReturnAsEnchantment = true
	ret.ReturnAsEnchantmentRiderSpan = riderSpan
	for _, index := range riderCandidates {
		sentences[index].ReturnAsEnchantmentRider = true
	}
}

// loneReturnToBattlefieldEffect returns the single return-to-battlefield effect
// across the sentences, or nil when the sentences hold zero or more than one
// such effect. Sibling effects of other kinds are permitted and ignored.
func loneReturnToBattlefieldEffect(sentences []Sentence) *EffectSyntax {
	var found *EffectSyntax
	for i := range sentences {
		for j := range sentences[i].Effects {
			if sentences[i].Effects[j].Kind != EffectReturn ||
				sentences[i].Effects[j].ToZone != zone.Battlefield {
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

// isChosenColorChooseTokens reports whether the sentence tokens are exactly
// "Choose a color" (with optional trailing periods). This bare color-choice
// sentence precedes "Add an amount of mana of that color equal to your devotion
// to that color." (Nykthos, Shrine to Nyx); the choice itself produces no
// standalone effect, so it is folded onto the chosen-color devotion add-mana.
func isChosenColorChooseTokens(tokens []shared.Token) bool {
	if !effectWordsAt(tokens, 0, "choose", "a", "color") {
		return false
	}
	rest := tokens[3:]
	for i := range rest {
		if rest[i].Kind != shared.Period {
			return false
		}
	}
	return true
}

// underOwnersControl reports the battlefield-destination ownership rider "under
// their owners' control" / "under its owner's control", under which each moved
// card enters controlled by its own owner rather than the resolving player. It
// is distinct from the "under your control" rider and the bare form.
func underOwnersControl(tokens []shared.Token) bool {
	words := normalizedWords(tokens)
	if !slices.Contains(words, "under") {
		return false
	}
	return effectContainsWords(words, "owners", "control") ||
		effectContainsWords(words, "owner's", "control")
}

// ability's lone chosen-color add-mana effect by widening that effect's span to
// cover the choice sentence, so the mana ability's coverage scan credits the
// choice. It succeeds only when the ability holds exactly one add-mana effect
// that carries a chosen-color body (devotion or dynamic count) and that effect is
// exact; otherwise it reports failure so the choice stays unrecognized and the
// card fails closed.
func creditChosenColorChoice(sentences []Sentence, chooseCandidates []int) bool {
	manaEffect := loneChosenColorManaEffect(sentences)
	if manaEffect == nil || !manaEffect.Exact {
		return false
	}
	for _, index := range chooseCandidates {
		if sentences[index].Span.Start.Offset < manaEffect.Span.Start.Offset {
			manaEffect.Span.Start = sentences[index].Span.Start
		}
	}
	return true
}

// loneChosenColorManaEffect returns the single chosen-color add-mana effect
// (devotion or dynamic count) across the sentences, or nil when the sentences
// hold zero or more than one such effect.
func loneChosenColorManaEffect(sentences []Sentence) *EffectSyntax {
	var found *EffectSyntax
	for i := range sentences {
		for j := range sentences[i].Effects {
			manaSyntax := sentences[i].Effects[j].Mana
			if sentences[i].Effects[j].Kind != EffectAddMana ||
				!manaSyntax.ChosenColorDevotion && !manaSyntax.ChosenColorDynamic {
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

// creditTokenCopyGrantRider folds a "[That token/It] gains <keyword>." rider
// sentence onto the sentences' lone create-copy-token effect. The created token
// gains the keyword(s); the rider sentence's effects are cleared and the
// sentence is marked so reference and coverage scans credit it to the create.
// It credits only when the sentences hold exactly one exact create-copy-token
// effect and a single matching gain-keyword sentence whose subject denotes the
// created token; otherwise nothing is folded and the card fails closed. It
// returns the folded sentence's legacy-effect and current-effect counts so the
// caller can correct its sequence-length bookkeeping.
func creditTokenCopyGrantRider(sentences []Sentence, atoms Atoms) (foldedLegacy, foldedEffects int, ok bool) {
	create := loneCopyTokenCreateEffect(sentences)
	if create == nil || !create.Exact {
		return 0, 0, false
	}
	for i := range sentences {
		if len(sentences[i].Effects) != 1 || sentences[i].Effects[0].Kind != EffectGain {
			continue
		}
		tokens := semanticEffectTokens(sentences[i].Tokens)
		keywords, match := tokenCopyGrantRiderKeywords(tokens, atoms)
		if !match {
			continue
		}
		create.TokenCopyGrantKeywords = keywords
		create.TokenCopyGrantRiderSpan = sentences[i].Span
		foldedEffects = len(sentences[i].Effects)
		if sentences[i].LegacyEffects {
			foldedLegacy = legacyEffectCount(tokens, atoms)
		}
		sentences[i].Effects = nil
		sentences[i].TokenCopyGrantRider = true
		return foldedLegacy, foldedEffects, true
	}
	return 0, 0, false
}

// creditCopyChooseNewTargetsRider folds the optional "You may choose new targets
// for the copy[ies]." rider sentence onto the ability's lone copy-stack-object
// effect: it sets CopyMayChooseNewTargets plus a coverage span on the copy and
// clears the rider sentence's effects so reference and coverage scans credit it.
// It credits only when the ability holds exactly one copy-stack-object effect,
// that copy is exact, and the rider sentence is exactly the recognized retarget
// clause; otherwise the rider stays uncredited and the card fails closed.
func creditCopyChooseNewTargetsRider(sentences []Sentence, atoms Atoms) (foldedLegacy, foldedEffects int, ok bool) {
	copyEffect := loneCopyStackObjectEffect(sentences)
	if copyEffect == nil || !copyEffect.Exact {
		return 0, 0, false
	}
	for i := range sentences {
		if len(sentences[i].Effects) != 1 || sentences[i].Effects[0].Kind != EffectChooseNewTargets {
			continue
		}
		tokens := semanticEffectTokens(sentences[i].Tokens)
		if !isCopyChooseNewTargetsRiderTokens(tokens) {
			continue
		}
		copyEffect.CopyMayChooseNewTargets = true
		copyEffect.CopyChooseNewTargetsRiderSpan = sentences[i].Span
		foldedEffects = len(sentences[i].Effects)
		if sentences[i].LegacyEffects {
			foldedLegacy = legacyEffectCount(tokens, atoms)
		}
		sentences[i].Effects = nil
		sentences[i].LegacyEffects = false
		sentences[i].CopyChooseNewTargetsRider = true
		return foldedLegacy, foldedEffects, true
	}
	return 0, 0, false
}

// loneCopyStackObjectEffect returns the single copy-stack-object effect across
// the sentences, or nil when the sentences hold zero or more than one.
func loneCopyStackObjectEffect(sentences []Sentence) *EffectSyntax {
	var found *EffectSyntax
	for i := range sentences {
		for j := range sentences[i].Effects {
			effect := &sentences[i].Effects[j]
			if effect.Kind != EffectCopyStackObject {
				continue
			}
			if found != nil {
				return nil
			}
			found = effect
		}
	}
	return found
}

// isCopyChooseNewTargetsRiderTokens reports whether the sentence tokens are
// exactly "You may choose new targets for the copy[ies]." The plural "copies"
// form covers multi-copy effects ("Copy ... X times. You may choose new targets
// for the copies."). Any other wording leaves the rider uncredited.
func isCopyChooseNewTargetsRiderTokens(tokens []shared.Token) bool {
	clause := strings.TrimSuffix(strings.ToLower(joinedEffectText(tokens)), ".")
	return clause == "you may choose new targets for the copy" ||
		clause == "you may choose new targets for the copies"
}

// loneCopyTokenCreateEffect returns the single create-copy-token effect across
// the sentences (a copy of a target, reference, or attached permanent), or nil
// when the sentences hold zero or more than one such effect.
func loneCopyTokenCreateEffect(sentences []Sentence) *EffectSyntax {
	var found *EffectSyntax
	for i := range sentences {
		for j := range sentences[i].Effects {
			effect := &sentences[i].Effects[j]
			if !effect.TokenCopyOfTarget && !effect.TokenCopyOfReference && !effect.TokenCopyOfAttached {
				continue
			}
			if found != nil {
				return nil
			}
			found = effect
		}
	}
	return found
}

// tokenCopyGrantRiderKeywords reports whether the sentence tokens are exactly
// "[That token/Those tokens/It/They] gain(s) <keyword>[ and <keyword> ...]." and
// returns the granted keyword kinds in source order. It fails closed for any
// trailing duration ("until end of turn"), quoted ability, or other content so
// only a plain keyword grant on the created token is folded.
func tokenCopyGrantRiderKeywords(tokens []shared.Token, atoms Atoms) ([]KeywordKind, bool) {
	verb := -1
	for i := range tokens {
		if equalWord(tokens[i], "gains") || equalWord(tokens[i], "gain") {
			verb = i
			break
		}
	}
	if verb <= 0 {
		return nil, false
	}
	subject := strings.ToLower(joinedEffectText(tokens[:verb]))
	switch subject {
	case "that token", "those tokens", "it", "they":
	default:
		return nil, false
	}
	keywordAtoms := atoms.KeywordsWithin(tokens)
	if len(keywordAtoms) == 0 {
		return nil, false
	}
	kinds := make([]KeywordKind, 0, len(keywordAtoms))
	texts := make([]string, 0, len(keywordAtoms))
	for _, keyword := range keywordAtoms {
		if keyword.Parameter.Kind != KeywordParameterNone {
			return nil, false
		}
		kinds = append(kinds, keyword.Kind)
		texts = append(texts, keyword.Text)
	}
	remainder := strings.TrimSuffix(joinedEffectText(tokens[verb+1:]), ".")
	expected := strings.Join(texts, " and ")
	if !strings.EqualFold(normalizeApostrophes(remainder), normalizeApostrophes(expected)) {
		return nil, false
	}
	return kinds, true
}

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

// isThisWayRegenerationRiderTokens reports whether the sentence tokens are a
// regeneration rider of the "destroyed this way" templated form, for example
// "A creature destroyed this way can't be regenerated." (Damn) or "Creatures
// destroyed this way can't be regenerated." Unlike the bare "that
// creature"/"those creatures" subject forms, this indefinite "destroyed this
// way" clause introduces no back-reference, so it contributes no phantom target
// or reference to the compiled effect and can fold onto the lone destroy safely.
// "Dealt damage this way" riders are intentionally excluded: they belong to a
// damage effect, which has no prevent-regeneration lowering yet, so they remain
// fail-closed instead of silently dropping the clause.
func isThisWayRegenerationRiderTokens(tokens []shared.Token) bool {
	end := len(tokens)
	for end > 0 && tokens[end-1].Kind == shared.Period {
		end--
	}
	core := tokens[:end]
	return endsWithWords(core, "destroyed", "this", "way", "can't", "be", "regenerated")
}

// endsWithWords reports whether the trailing tokens match words in order.
func endsWithWords(tokens []shared.Token, words ...string) bool {
	return effectWordsAt(tokens, len(tokens)-len(words), words...)
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

// trailingDynamicCountInClause reports whether amount is a trailing dynamic
// count phrase ("for each ...", "equal to ...", "where X is ...") whose tokens
// fall inside clause. A leading count prefix ("For each X, create ...") lives
// before the verb, so its span starts before the clause and is excluded here.
func trailingDynamicCountInClause(clause []shared.Token, amount EffectAmountSyntax) bool {
	switch amount.DynamicForm {
	case EffectDynamicAmountFormForEach, EffectDynamicAmountFormEqual, EffectDynamicAmountFormWhereX:
	default:
		return false
	}
	if len(clause) == 0 {
		return false
	}
	return amount.Span.Start.Offset >= clause[0].Span.Start.Offset
}

// stripLeadingConditionClause drops a leading "As long as this card/creature is
// in your graveyard and ..." condition clause so the subject grammar sees only
// the effect's group subject ("creatures you control"). The first effect's
// ownership tokens begin at the sentence start, so the graveyard zone-of-
// function condition would otherwise prevent the group subject from being
// recognized at token zero. The strip is restricted to graveyard conditions so
// other leading conditions keep their existing recognition path unchanged.
func stripLeadingConditionClause(tokens []shared.Token) []shared.Token {
	if len(tokens) == 0 {
		return tokens
	}
	intro, width := conditionIntroAt(tokens, 0)
	if intro == ConditionIntroUnknown {
		return tokens
	}
	body := tokens[width:]
	if _, ok := cutTokenPrefix(body, "this", "card", "is", "in", "your", "graveyard", "and"); !ok {
		if _, ok := cutTokenPrefix(body, "this", "creature", "is", "in", "your", "graveyard", "and"); !ok {
			return tokens
		}
	}
	end := conditionClauseEnd(tokens, 0)
	if end < len(tokens) && tokens[end].Kind == shared.Comma {
		return tokens[end+1:]
	}
	return tokens
}

// stripLeadingDurationClause removes a sentence-leading duration clause
// ("Until end of turn, ...", "Until your next turn, ...") from the token slice,
// returning the remaining tokens and the duration it names. A duration stated
// once at the front of a sentence applies to every continuous effect the
// sentence produces (CR 611.2: "Until end of turn, creatures you control gain
// trample and get +X/+X ..."), so the parser lifts it off the front where it
// would otherwise derail the group-subject parse and leave the trailing effect
// with no duration.
func stripLeadingDurationClause(tokens []shared.Token, atoms Atoms) ([]shared.Token, EffectDurationKind) {
	if len(tokens) == 0 || !equalWord(tokens[0], "until") {
		return tokens, EffectDurationNone
	}
	comma := -1
	for i := range tokens {
		if tokens[i].Kind == shared.Comma {
			comma = i
			break
		}
	}
	if comma <= 0 {
		return tokens, EffectDurationNone
	}
	duration := parseEffectDuration(tokens[:comma], atoms)
	if duration == EffectDurationNone {
		return tokens, EffectDurationNone
	}
	return tokens[comma+1:], duration
}

// parseSpecialEffects dispatches the sentence to the whole-sentence effect
// recognizers that bypass the per-clause loop in parseEffects. It returns the
// first recognizer's result, or ok=false when none match and the general
// per-clause parsing should run. Order is significant and matches the original
// dispatch sequence.
func parseSpecialEffects(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	for _, recognize := range []func() ([]EffectSyntax, bool){
		func() ([]EffectSyntax, bool) { return parsePassiveTokenDoublingEffects(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseEntersAsCopyEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseBecomeCopyEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseDrawEmptyLibraryWinReplacement(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseDrawDoublingReplacement(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseLifeGainReplacement(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parsePunisherEachLoseLifeEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseLibraryTopReorderEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseGroupEntersTappedEffect(sentence, tokens) },
		func() ([]EffectSyntax, bool) { return parsePlayerProtectionEffects(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseGroupPhaseOutEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseMassReanimationExchangeEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseAdditionalLandPlaysEffect(sentence, tokens, atoms) },
		func() ([]EffectSyntax, bool) { return parseCastAsThoughFlashEffect(sentence, tokens) },
		func() ([]EffectSyntax, bool) { return parseCantCastSpellsEffect(sentence, tokens) },
		func() ([]EffectSyntax, bool) { return parseGroupMustAttackEffect(sentence, tokens) },
		func() ([]EffectSyntax, bool) { return parseSpellsCantBeCounteredEffect(sentence, tokens) },
	} {
		if effects, ok := recognize(); ok {
			return effects, true
		}
	}
	return nil, false
}

func parseEffects(sentence Sentence, tokens []shared.Token, atoms Atoms) []EffectSyntax {
	if effects, ok := parseSpecialEffects(sentence, tokens, atoms); ok {
		return effects
	}
	if effects, ok := parsePreventCombatDamageEffect(sentence, tokens, atoms); ok {
		return effects
	}
	indices := effectIndices(tokens, atoms)
	requiresOrderedLowering := legacyEffectCount(tokens, atoms) > 1
	effects := make([]EffectSyntax, 0, len(indices))
	_, leadingDuration := stripLeadingDurationClause(tokens, atoms)
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
		subjectTokens := stripLeadingConditionClause(ownership)
		subjectTokens, _ = stripLeadingDurationClause(subjectTokens, atoms)
		staticSubject := parseEffectStaticSubject(subjectTokens, atoms)
		payment := parseEffectPayment(tokens, atoms)
		connection, connectionSpan := effectConnection(tokens, indices, effectIndex)
		optional, optionalSpan := effectOptional(tokens, tokenIndex)
		context := effectContextAt(tokens, tokenIndex, atoms)
		if effectIndex > 0 && !effectHasExplicitSubject(tokens, tokenIndex, atoms.SelfNameSpans()) &&
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
		duration := parseEffectDuration(durationTokens, atoms)
		if duration == EffectDurationNone {
			duration = leadingDuration
		}
		kind := effectKindAt(tokens, tokenIndex)
		if loseGameObject(kind, clause) {
			kind = EffectLoseGame
		}
		entersColorChoice, entersColorChoiceExclude := entersColorChoiceSyntax(kind, clause)
		doublePower, doubleToughness := false, false
		if kind == EffectDouble {
			if object, okDouble := parseDoublePTObject(clause, atoms); okDouble {
				staticSubject = object.Subject
				doublePower, doubleToughness = object.DoublePower, object.DoubleToughness
			}
		}
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
		// A create-token clause whose amount is a trailing "for each <permanent>
		// you control" (or "equal to ...") count phrase ("Create a 0/1 green
		// Plant creature token for each land you control.") embeds the
		// counted-subject selector in the same clause as the token's own
		// characteristics. Like the deal-damage case above, scope the token
		// Selection to the run of tokens before the count phrase so the count
		// subject's filters do not fold into the token's type line.
		selectionClause := clause
		switch {
		case kind == EffectDealDamage && amount.DynamicForm == EffectDynamicAmountFormWhereX:
			selectionClause = tokensBeforeOffset(clause, amount.Span.Start.Offset)
		case kind == EffectCreate && trailingDynamicCountInClause(clause, amount):
			selectionClause = tokensBeforeOffset(clause, amount.Span.Start.Offset)
		case kind == EffectMill && trailingDynamicCountInClause(clause, amount):
			// "mills cards equal to <dynamic>" embeds the count subject in the
			// same clause as the milled-card noun. Scoping the selection to the
			// tokens before the count phrase keeps a competing permanent noun in
			// the amount (e.g. "the sacrificed creature's power") from folding
			// into the milled-card selection.
			selectionClause = tokensBeforeOffset(clause, amount.Span.Start.Offset)
		default:
		}
		eachSourceDamageGroup, eachSourceDamageRecipient := eachSourceDamageSyntax(kind, tokens[ownershipStart:tokenIndex], clause, atoms)
		fallbackOnInability := effectFallbackOnInability(tokens, ownershipStart, tokenIndex)
		effects = append(effects, EffectSyntax{
			Kind:                      kind,
			Context:                   context,
			Connection:                connection,
			ConnectionSpan:            connectionSpan,
			Span:                      sentence.Span,
			VerbSpan:                  tokens[tokenIndex].Span,
			ClauseSpan:                ownershipSpan,
			Text:                      sentence.Text,
			Tokens:                    append([]shared.Token(nil), ownership...),
			Duration:                  duration,
			DelayedTiming:             delayed,
			Selection:                 parseSelection(selectionClause, atoms),
			DamageRecipientPair:       parseDamageRecipientPair(kind, clause, amount, atoms),
			EachSourceDamageGroup:     eachSourceDamageGroup,
			EachSourceDamageRecipient: eachSourceDamageRecipient,
			Amount:                    amount,
			PowerDelta:                power,
			ToughnessDelta:            toughness,
			TokenPower:                tokenPower,
			TokenToughness:            tokenToughness,
			TokenPTKnown:              tokenPTKnown,
			TokenKeywords:             parseTokenKeywords(kind, clause, atoms),
			TokenName:                 parseTokenName(kind, clause),
			TokenChoice:               parseTokenChoice(kind, clause),
			StaticSubject:             staticSubject,
			DoublePower:               doublePower,
			DoubleToughness:           doubleToughness,
			CounterKind:               counterKind,
			CounterKnown:              counterKnown,
			CounterRecipientAttached:  counterRecipientAttached(kind, counterKnown, clause),
			MoveCountersAll:           kind == EffectMoveCounters && moveAllCountersClause(clause),
			FromZone:                  firstZone(atoms, span, ZoneRoleFrom),
			ToZone:                    toZone,
			Destination:               parseEffectDestination(ownership),
			EntersTapped:              effectWordsAtAny(ownership, "battlefield", "tapped"),
			EntersTappedSelf:          entersTappedSelfSyntax(kind, clause),
			EntersColorChoice:         entersColorChoice,
			EntersColorChoiceExclude:  entersColorChoiceExclude,
			EntersTypeChoice:          entersTypeChoiceSyntax(kind, clause),
			EntersWithCounters:        entersWithCountersSyntax(kind, clause),
			UnderYourControl:          effectContainsWords(normalizedWords(ownership), "under", "your", "control"),
			UnderOwnersControl:        underOwnersControl(ownership),
			CastAsAdventure:           effectContainsWords(normalizedWords(clause), "as", "an", "adventure"),
			CastWithoutPayingManaCost: kind == EffectCast &&
				effectContainsWords(normalizedWords(clause), "without", "paying", "its", "mana", "cost"),
			Negated:                 effectIsNegated(tokens, tokenIndex) && !fallbackOnInability,
			FallbackOnInability:     fallbackOnInability,
			Optional:                optional,
			OptionalSpan:            optionalSpan,
			LifeObject:              gainLoseLifeObject(kind, clause),
			Symbol:                  firstEffectSymbol(clause),
			Mana:                    parseEffectMana(kind, clause, nextConnection != EffectConnectionNone),
			Replacement:             parseEffectReplacement(ownership, atoms),
			References:              referencesInSpan(atoms, ownershipSpan),
			SubjectReferences:       referencesInSpan(atoms, shared.SpanOf(tokens[ownershipStart:tokenIndex])),
			Targets:                 targetsInSpan(sentence.Targets, ownershipSpan),
			SubjectTargets:          targetsInSpan(sentence.Targets, shared.SpanOf(tokens[ownershipStart:tokenIndex])),
			Payment:                 payment,
			RequiresOrderedLowering: requiresOrderedLowering,
		})
	}

	for i := range effects {
		finalizeParsedEffect(&effects[i], sentence, atoms)
	}
	return effects
}

func finalizeParsedEffect(effect *EffectSyntax, sentence Sentence, atoms Atoms) {
	effect.Divided = dividedDamageEffect(effect)
	effect.DamageRecipientReference = damageRecipientReference(effect)
	effect.SelfDamageRiderValue, effect.HasSelfDamageRider = damageSelfRider(effect)
	effect.TargetControllerDamageRiderValue, effect.TargetControllerDamageRiderRecipient = damageTargetControllerRider(effect)
	effect.SecondTargetDamageRiderValue, effect.HasSecondTargetDamageRider = damageSecondTargetRider(effect)
	effect.Dig = parseDigPut(effect)
	effect.HandLibraryPut = parseHandLibraryPut(effect)
	effect.HandDiscard = parseHandDiscard(effect)
	effect.DiscardEntireHand = parseDiscardEntireHand(effect)
	effect.SearchSplit = parseSearchSplitPut(effect)
	effect.GraveyardZoneExile = parseGraveyardZoneExile(effect)
	effect.Additional = drawAdditionalCardsQualifier(effect)
	effect.Exact = exactEffectSyntax(effect)
	if recognizeTargetOpponentHandMana(effect) {
		effect.Exact = true
	}
	if recognizeDynamicCountMana(effect) {
		effect.Exact = true
	}
	if recognizeColorsAmongControlledMana(effect, atoms) {
		effect.Exact = true
	}
	if recognizeEachColorAmongControlledMana(effect, atoms) {
		effect.Exact = true
	}
	if recognizeTargetColorIfRider(effect, atoms) {
		effect.Exact = true
	}
	// "Destroy each <permanent group>" selects every matching permanent like the
	// plural "all" form, so flag its selection as a mass group to lower to a
	// battlefield-group destroy. Scoped to the recognized destroy mass form so
	// "each creature" damage recipients and "each player" distributive effects on
	// other effect kinds are untouched.
	if effect.Kind == EffectDestroy && exactMassEachEffectSyntax(effect, "Destroy each ") {
		effect.Selection.All = true
	}
	effect.TokenCopyOfTarget = exactCreateCopyTokenEffectSyntax(effect)
	effect.TokenCopyOfReference = exactCreateCopyTokenReferenceEffectSyntax(effect)
	effect.TokenCopyOfAttached = exactCreateCopyTokenAttachedEffectSyntax(effect)
	if group, ok := exactCreateCopyTokenForEachEffectSyntax(effect, atoms); ok {
		effect.TokenCopyOfForEach = true
		effect.TokenCopyForEachGroup = group
	}
	effect.Mana.LegacyBodyExact = legacyExactManaBody(effect, sentence)
	if effect.Kind == EffectSearch {
		effect.UnsupportedDetail = searchUnsupportedDetail(effect)
		effect.SearchSharedSubtype = searchSharedSubtypeRider(effect)
		effect.SearchDestination = searchDestinationPosition(effect)
	}
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

// parsePassiveTokenDoublingEffects recognizes the passive-voice token-doubling
// replacement "If one or more tokens would be created under your control, twice
// that many of those tokens are created instead." (Mondrak, Adrix and Nev). Its
// active-voice equivalent "If an effect would create one or more tokens under
// your control, it creates twice that many of those tokens instead." (Doubling
// Season, Anointed Procession, Parallel Lives) parses through the ordinary
// create-verb path. The passive wording carries no active "create" verb, so it
// is recognized here and emitted as the same two EffectCreate instructions the
// active form produces: the would-create group and the doubled output marked
// EffectReplacementTwiceThatMany. The matching intervening-if condition is
// recognized separately by recognizeTokenCreationCondition.
func parsePassiveTokenDoublingEffects(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	commaIndex, anyController, ok := matchPassiveTokenDoubling(tokens)
	if !ok {
		return nil, false
	}
	condition := tokens[:commaIndex]
	resolving := tokens[commaIndex+1:]
	createdIndex := commaIndex - 4
	if anyController {
		createdIndex = commaIndex - 1
	}
	createEffect := EffectSyntax{
		Kind:             EffectCreate,
		Context:          EffectContextController,
		Span:             shared.SpanOf(condition),
		VerbSpan:         tokens[createdIndex].Span,
		ClauseSpan:       shared.SpanOf(condition),
		Text:             sentence.Text,
		Tokens:           append([]shared.Token(nil), condition...),
		Amount:           EffectAmountSyntax{Value: 1, Known: true},
		UnderYourControl: !anyController,
	}
	doubledIndex := commaIndex + 8
	doubledEffect := EffectSyntax{
		Kind:       EffectCreate,
		Context:    EffectContextReferencedObject,
		Span:       shared.SpanOf(resolving),
		VerbSpan:   tokens[doubledIndex].Span,
		ClauseSpan: shared.SpanOf(resolving),
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), resolving...),
		Amount:     EffectAmountSyntax{Value: 2, Known: true},
		Replacement: EffectReplacementSyntax{
			Kind: EffectReplacementTwiceThatMany,
			Span: tokens[len(tokens)-2].Span,
		},
		References: referencesInSpan(atoms, shared.SpanOf(resolving)),
	}
	return []EffectSyntax{createEffect, doubledEffect}, true
}

// matchPassiveTokenDoubling reports the index of the comma separating the
// would-create condition clause from the doubled output clause when tokens spell
// the passive token-doubling replacement. The controller-only wording ("...under
// your control, ...", Mondrak) and the controller-agnostic wording ("...would be
// created, ...", Primal Vigor) are distinguished by anyController, which the
// any-player runtime node honors via an any-player scope.
func matchPassiveTokenDoubling(tokens []shared.Token) (commaIndex int, anyController, ok bool) {
	if len(tokens) == 22 &&
		effectWordsAt(tokens, 0, "if", "one", "or", "more", "tokens", "would", "be", "created") &&
		effectWordsAt(tokens, 8, "under", "your", "control") &&
		tokens[11].Kind == shared.Comma &&
		effectWordsAt(tokens, 12, "twice", "that", "many", "of", "those", "tokens", "are", "created", "instead") &&
		tokens[21].Kind == shared.Period {
		return 11, false, true
	}
	if len(tokens) == 19 &&
		effectWordsAt(tokens, 0, "if", "one", "or", "more", "tokens", "would", "be", "created") &&
		tokens[8].Kind == shared.Comma &&
		effectWordsAt(tokens, 9, "twice", "that", "many", "of", "those", "tokens", "are", "created", "instead") &&
		tokens[18].Kind == shared.Period {
		return 8, true, true
	}
	return 0, false, false
}

// parseDrawEmptyLibraryWinReplacement recognizes the draw-from-empty-library win
// replacement "If you would draw a card while your library has no cards in it,
// you win the game instead." (Laboratory Maniac, Jace, Wielder of Mysteries) and
// emits a single win-the-game effect for the resolving clause. The matching
// intervening-if condition is recognized separately by
// recognizeDrawFromEmptyLibraryCondition; the runtime replacement is otherwise
// self-contained, so the would-draw clause needs no effect of its own.
func parseDrawEmptyLibraryWinReplacement(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	commaIndex, ok := matchDrawEmptyLibraryWin(tokens)
	if !ok {
		return nil, false
	}
	resolving := tokens[commaIndex+1:]
	winIndex := commaIndex + 2
	return []EffectSyntax{{
		Kind:       EffectWinGame,
		Context:    EffectContextController,
		Span:       shared.SpanOf(tokens),
		VerbSpan:   tokens[winIndex].Span,
		ClauseSpan: shared.SpanOf(tokens),
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), resolving...),
		Replacement: EffectReplacementSyntax{
			Kind: EffectReplacementInstead,
			Span: tokens[len(tokens)-2].Span,
		},
		References: referencesInSpan(atoms, shared.SpanOf(resolving)),
		Exact:      true,
	}}, true
}

// parsePunisherEachLoseLifeEffect recognizes the "punisher" family ("<group>
// punisherUnlessClauseAt reports whether tokens[i:] begins the punisher idiom
// "unless that player sacrifices ... [or discards ...]". This per-player
// alternative-cost clause is consumed wholesale by parsePunisherEachLoseLifeEffect
// and must not be parsed as a game-state condition. It is distinguished from the
// "unless that player pays ..." payment idiom by the sacrifice/discard verb.
func punisherUnlessClauseAt(tokens []shared.Token, i int) bool {
	if !effectWordsAt(tokens, i, "unless", "that", "player") || i+3 >= len(tokens) {
		return false
	}
	return equalWord(tokens[i+3], "sacrifices") || equalWord(tokens[i+3], "discards")
}

// loses N life unless that player sacrifices a <permanent> [of their choice]
// [or discards a card].") as a single EffectPunisherLoseLife effect. The group
// must be each-opponent / each-player / each-other-player; the alternatives are
// a filtered sacrifice and/or a discard-a-card, joined by "or" in either order.
func parsePunisherEachLoseLifeEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	verbIndex := -1
	for i, token := range tokens {
		if equalWord(token, "loses") || equalWord(token, "lose") {
			verbIndex = i
			break
		}
	}
	if verbIndex < 0 || verbIndex+2 >= len(tokens) {
		return nil, false
	}
	context := effectContextAt(tokens, verbIndex, atoms)
	switch context {
	case EffectContextEachOpponent, EffectContextEachPlayer, EffectContextEachOtherPlayer:
	default:
		return nil, false
	}
	amount, ok := effectNumber(tokens[verbIndex+1], atoms)
	if !ok || amount < 1 || !equalWord(tokens[verbIndex+2], "life") {
		return nil, false
	}
	rest := tokens[verbIndex+3:]
	if len(rest) < 4 ||
		!equalWord(rest[0], "unless") ||
		!equalWord(rest[1], "that") ||
		!equalWord(rest[2], "player") {
		return nil, false
	}
	options := rest[3:]
	if n := len(options); n > 0 && options[n-1].Kind == shared.Period {
		options = options[:n-1]
	}
	segments := splitPunisherOptions(options)
	if len(segments) == 0 || len(segments) > 2 {
		return nil, false
	}
	effect := EffectSyntax{
		Kind:       EffectPunisherLoseLife,
		Context:    context,
		Span:       shared.SpanOf(tokens),
		VerbSpan:   tokens[verbIndex].Span,
		ClauseSpan: shared.SpanOf(tokens),
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Amount:     EffectAmountSyntax{Value: amount, Known: true, Span: tokens[verbIndex+1].Span},
		Exact:      true,
	}
	for _, segment := range segments {
		if len(segment) == 0 {
			return nil, false
		}
		switch {
		case equalWord(segment[0], "sacrifices") || equalWord(segment[0], "sacrifice"):
			if effect.PunisherSacrifice {
				return nil, false
			}
			selectionTokens := stripOfTheirChoice(segment[1:])
			if len(selectionTokens) == 0 {
				return nil, false
			}
			selection := parseSelection(selectionTokens, atoms)
			effect.PunisherSacrifice = true
			effect.Selection = selection
		case equalWord(segment[0], "discards") || equalWord(segment[0], "discard"):
			if effect.PunisherDiscard || !punisherDiscardACard(segment[1:]) {
				return nil, false
			}
			effect.PunisherDiscard = true
		default:
			return nil, false
		}
	}
	if !effect.PunisherSacrifice && !effect.PunisherDiscard {
		return nil, false
	}
	return []EffectSyntax{effect}, true
}

// splitPunisherOptions splits a punisher's avoidance clause on its top-level
// "or" connectives, returning each option's tokens.
func splitPunisherOptions(tokens []shared.Token) [][]shared.Token {
	var segments [][]shared.Token
	start := 0
	for i, token := range tokens {
		if equalWord(token, "or") {
			segments = append(segments, tokens[start:i])
			start = i + 1
		}
	}
	return append(segments, tokens[start:])
}

// stripOfTheirChoice drops a trailing "of their choice" qualifier from a
// sacrifice selection's tokens ("a nonland permanent of their choice").
func stripOfTheirChoice(tokens []shared.Token) []shared.Token {
	if n := len(tokens); n >= 3 &&
		equalWord(tokens[n-3], "of") &&
		equalWord(tokens[n-2], "their") &&
		equalWord(tokens[n-1], "choice") {
		return tokens[:n-3]
	}
	return tokens
}

// punisherDiscardACard reports whether tokens spell the "a card" object of a
// punisher's discard alternative.
func punisherDiscardACard(tokens []shared.Token) bool {
	return len(tokens) == 2 && equalWord(tokens[0], "a") && equalWord(tokens[1], "card")
}

// matchDrawEmptyLibraryWin reports the index of the comma separating the
// would-draw condition clause from the win-the-game result when tokens spell the
// draw-from-empty-library win replacement.
func matchDrawEmptyLibraryWin(tokens []shared.Token) (int, bool) {
	if len(tokens) != 21 ||
		!effectWordsAt(tokens, 0, "if", "you", "would", "draw", "a", "card") ||
		!effectWordsAt(tokens, 6, "while", "your", "library", "has", "no", "cards", "in", "it") ||
		tokens[14].Kind != shared.Comma ||
		!effectWordsAt(tokens, 15, "you", "win", "the", "game", "instead") ||
		tokens[20].Kind != shared.Period {
		return 0, false
	}
	return 14, true
}

// parseDrawDoublingReplacement recognizes the draw-doubling replacement "If you
// would draw a card[ except the first one you draw in each of your draw steps],
// draw <N> cards instead." (Thought Reflection, Teferi's Ageless Insight) and
// emits a single draw effect carrying the replacement multiplier N for the
// resolving clause. The matching intervening-if condition is recognized
// separately by recognizeDrawCardReplacementCondition.
func parseDrawDoublingReplacement(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	commaIndex, amount, ok := matchDrawDoubling(tokens)
	if !ok {
		return nil, false
	}
	resolving := tokens[commaIndex+1:]
	drawIndex := commaIndex + 1
	return []EffectSyntax{{
		Kind:       EffectDraw,
		Context:    EffectContextController,
		Span:       shared.SpanOf(tokens),
		VerbSpan:   tokens[drawIndex].Span,
		ClauseSpan: shared.SpanOf(tokens),
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), resolving...),
		Amount:     EffectAmountSyntax{Value: amount, Known: true},
		Replacement: EffectReplacementSyntax{
			Kind: EffectReplacementInstead,
			Span: tokens[len(tokens)-2].Span,
		},
		References: referencesInSpan(atoms, shared.SpanOf(resolving)),
		Exact:      true,
	}}, true
}

// matchDrawDoubling reports the comma index separating the would-draw condition
// from the "draw <N> cards instead" result and the multiplier N when tokens
// spell a draw-doubling replacement. The condition must be the plain would-draw
// or the draw-step exception form, and the result an N (>= 2) card draw.
func matchDrawDoubling(tokens []shared.Token) (commaIndex, amount int, ok bool) {
	if len(tokens) < 6 || tokens[len(tokens)-1].Kind != shared.Period {
		return 0, 0, false
	}
	if !effectWordsAt(tokens, 0, "if", "you", "would", "draw", "a", "card") {
		return 0, 0, false
	}
	comma := -1
	for i := range tokens {
		if tokens[i].Kind == shared.Comma {
			comma = i
			break
		}
	}
	if comma < 0 || !drawDoublingConditionBody(tokens[1:comma]) {
		return 0, 0, false
	}
	result := tokens[comma+1 : len(tokens)-1]
	if len(result) != 4 ||
		!equalWord(result[0], "draw") ||
		!equalWord(result[2], "cards") ||
		!equalWord(result[3], "instead") {
		return 0, 0, false
	}
	n, valueOK := CardinalWordValue(result[1].Text)
	if !valueOK || n < 2 {
		return 0, 0, false
	}
	return comma, n, true
}

// drawDoublingConditionBody reports whether a draw-doubling condition body is one
// of the two supported forms: the plain "you would draw a card" or the draw-step
// exception "you would draw a card except the first one you draw in each of your
// draw steps".
func drawDoublingConditionBody(body []shared.Token) bool {
	if tokenWordsEqual(body, "you", "would", "draw", "a", "card") {
		return true
	}
	return tokenWordsEqual(body,
		"you", "would", "draw", "a", "card",
		"except", "the", "first", "one", "you", "draw",
		"in", "each", "of", "your", "draw", "steps")
}

// parseLifeGainReplacement recognizes the life-gain replacement "If you would
// gain life, you gain twice that much life instead." (multiplier two) and "If
// you would gain life, you gain that much life plus N instead." (additive
// bonus), emitting a single gain effect carrying the replacement so the
// would-gain condition clause does not become a spurious effect of its own. The
// matching intervening-if condition is recognized separately by
// recognizeLifeGainCondition.
func parseLifeGainReplacement(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	commaIndex, replacement, ok := matchLifeGainReplacement(tokens, atoms)
	if !ok {
		return nil, false
	}
	resolving := tokens[commaIndex+1:]
	gainIndex := commaIndex + 1
	replacement.Span = tokens[len(tokens)-2].Span
	return []EffectSyntax{{
		Kind:        EffectGain,
		Context:     EffectContextController,
		Span:        shared.SpanOf(tokens),
		VerbSpan:    tokens[gainIndex].Span,
		ClauseSpan:  shared.SpanOf(tokens),
		Text:        sentence.Text,
		Tokens:      append([]shared.Token(nil), resolving...),
		Replacement: replacement,
		References:  referencesInSpan(atoms, shared.SpanOf(resolving)),
		Exact:       true,
	}}, true
}

// matchLifeGainReplacement reports the comma index separating the would-gain
// condition from its "you gain ... life instead" result and the replacement
// (twice-that-much or that-much-plus-N) when tokens spell a life-gain
// replacement.
func matchLifeGainReplacement(tokens []shared.Token, atoms Atoms) (commaIndex int, replacement EffectReplacementSyntax, ok bool) {
	if len(tokens) < 6 || tokens[len(tokens)-1].Kind != shared.Period {
		return 0, EffectReplacementSyntax{}, false
	}
	if !effectWordsAt(tokens, 0, "if") {
		return 0, EffectReplacementSyntax{}, false
	}
	comma := -1
	for i := range tokens {
		if tokens[i].Kind == shared.Comma {
			comma = i
			break
		}
	}
	if comma < 0 || !tokenWordsEqual(tokens[1:comma], "you", "would", "gain", "life") {
		return 0, EffectReplacementSyntax{}, false
	}
	result := tokens[comma+1 : len(tokens)-1]
	if tokenWordsEqual(result, "you", "gain", "twice", "that", "much", "life", "instead") {
		return comma, EffectReplacementSyntax{Kind: EffectReplacementTwiceThatMuch}, true
	}
	if len(result) == 8 &&
		tokenWordsEqual(result[:6], "you", "gain", "that", "much", "life", "plus") &&
		equalWord(result[7], "instead") {
		if n, valueOK := effectNumber(result[6], atoms); valueOK && n > 0 {
			return comma, EffectReplacementSyntax{Kind: EffectReplacementThatMuchPlus, Amount: n}, true
		}
	}
	return 0, EffectReplacementSyntax{}, false
}

func recognizeImpulseExileSequence(sentences []Sentence) bool {
	if len(sentences) != 2 {
		return false
	}
	amount, ok := matchImpulseExileClause(strings.TrimSpace(sentences[0].Text))
	if !ok {
		return false
	}
	duration, ok := matchImpulsePlayPermissionClause(strings.TrimSpace(sentences[1].Text), amount)
	if !ok {
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
		Amount:     EffectAmountSyntax{Value: amount, Known: true},
		Duration:   duration,
		Exact:      true,
	}}
	return true
}

// matchImpulseExileClause recognizes "Exile the top card of your library." and
// its counted plural "Exile the top <N> cards of your library." (N a cardinal
// word two..ten), returning the fixed number of cards exiled.
func matchImpulseExileClause(text string) (int, bool) {
	if strings.EqualFold(text, "Exile the top card of your library.") {
		return 1, true
	}
	const prefix = "Exile the top "
	const suffix = " cards of your library."
	if len(text) <= len(prefix)+len(suffix) ||
		!strings.EqualFold(text[:len(prefix)], prefix) ||
		!strings.EqualFold(text[len(text)-len(suffix):], suffix) {
		return 0, false
	}
	count, ok := CardinalWordValue(text[len(prefix) : len(text)-len(suffix)])
	if !ok || count < 2 {
		return 0, false
	}
	return count, true
}

// matchImpulsePlayPermissionClause recognizes the temporary play-permission
// sentence that follows an impulse exile: "You may play <object> this turn.",
// the "until end of turn" variants (trailing or leading "Until end of turn,"),
// where <object> agrees in number with the count exiled ("it"/"that card" for a
// single card, "them"/"those cards" for several). It returns the play window.
func matchImpulsePlayPermissionClause(text string, amount int) (EffectDurationKind, bool) {
	for _, object := range impulsePlayObjects(amount) {
		switch {
		case strings.EqualFold(text, "You may play "+object+" this turn."):
			return EffectDurationThisTurn, true
		case strings.EqualFold(text, "You may play "+object+" until end of turn."),
			strings.EqualFold(text, "Until end of turn, you may play "+object+"."):
			return EffectDurationUntilEndOfTurn, true
		case strings.EqualFold(text, "You may play "+object+" until the end of your next turn."),
			strings.EqualFold(text, "Until the end of your next turn, you may play "+object+"."):
			return EffectDurationUntilEndOfYourNextTurn, true
		}
	}
	return EffectDurationNone, false
}

// impulsePlayObjects lists the demonstratives an impulse play-permission sentence
// uses to refer back to the exiled cards, matching grammatical number to the
// count exiled.
func impulsePlayObjects(amount int) []string {
	if amount == 1 {
		return []string{"it", "that card"}
	}
	return []string{"them", "those cards"}
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

// recognizeDynamicCountMana types an add-mana body whose produced amount scales
// with a dynamic count: a fixed-color battlefield count ("for each <permanent>
// you control", recognizeControlledCountMana), a chosen-color battlefield count
// ("equal to <count>", recognizeChosenColorCountMana), or a source-counter count
// ("for each <kind> counter on this permanent", recognizeSourceCounterCountMana).
func recognizeDynamicCountMana(effect *EffectSyntax) bool {
	return recognizeControlledCountMana(effect) ||
		recognizeChosenColorCountMana(effect) ||
		recognizeSourceCounterCountMana(effect) ||
		recognizeAnyOneColorDynamicMana(effect)
}

// recognizeTargetColorIfRider reports whether a single-target counter or destroy
// effect carries the trailing color rider "if it's <color>" / "if it is
// <color>" (Pyroblast, Red Elemental Blast). The rider is segmented separately
// as a ConditionPredicateTargetColor condition (which carries the color and is
// consumed by counter/destroy lowering), but it leaves the verb clause text
// reading "<verb> <target> if it's <color>." so exactEffectSyntax cannot
// round-trip the bare "<verb> <target>." shape. This credits the effect as exact
// when the only trailing text past the lone exact target is the recognized color
// rider, so lowering can attach the resolving target-color gate.
func recognizeTargetColorIfRider(effect *EffectSyntax, atoms Atoms) bool {
	if effect.Kind != EffectCounter && effect.Kind != EffectDestroy {
		return false
	}
	if len(effect.Targets) != 1 ||
		!effect.Targets[0].Exact ||
		effect.Targets[0].Cardinality.Min != 1 ||
		effect.Targets[0].Cardinality.Max != 1 {
		return false
	}
	targetEnd := effect.Targets[0].Span.End.Offset
	var rider []shared.Token
	for _, token := range effect.Tokens {
		if token.Span.Start.Offset >= targetEnd {
			rider = append(rider, token)
		}
	}
	for len(rider) > 0 && rider[len(rider)-1].Kind == shared.Period {
		rider = rider[:len(rider)-1]
	}
	rest, ok := cutTokenPrefix(rider, "if", "it's")
	if !ok {
		if rest, ok = cutTokenPrefix(rider, "if", "it", "is"); !ok {
			return false
		}
	}
	if len(rest) != 1 {
		return false
	}
	_, ok = atoms.ColorAt(rest[0].Span)
	return ok
}

// recognizeAnyOneColorDynamicMana types the add-mana body "X mana of any one
// color" or "an amount of mana of any one color" (Kami of Whispered Hopes:
// "Add X mana of any one color, where X is this creature's power."). The
// produced quantity is the dynamic amount already typed onto effect.Amount by
// parseEffectAmount ("where X is <dynamic>" or "equal to <dynamic>"); the body
// itself is left unrecognized by parseEffectMana. This credits the
// freely-chosen single-color output when the body matches exactly and a dynamic
// amount is present, so the lowerer can produce that many mana of one chosen
// color. It fails closed when no dynamic amount is present.
func recognizeAnyOneColorDynamicMana(effect *EffectSyntax) bool {
	if effect.Kind != EffectAddMana ||
		effect.Amount.DynamicKind == EffectDynamicAmountNone {
		return false
	}
	switch effect.Amount.DynamicForm {
	case EffectDynamicAmountFormWhereX, EffectDynamicAmountFormEqual:
	default:
		return false
	}
	body := manaBodyBeforeAmount(effect)
	for len(body) > 0 && body[len(body)-1].Kind == shared.Comma {
		body = body[:len(body)-1]
	}
	if !anyOneColorDynamicManaBody(body) {
		return false
	}
	effect.Mana = EffectManaSyntax{Span: shared.SpanOf(body), AnyOneColorDynamic: true}
	return true
}

// anyOneColorDynamicManaBody reports whether body is the leading add-mana phrase
// of an "any one color" dynamic-amount output: "X mana of any one color" or "an
// amount of mana of any one color".
func anyOneColorDynamicManaBody(body []shared.Token) bool {
	if len(body) == 6 &&
		equalWord(body[0], "x") &&
		effectWordsAt(body, 1, "mana", "of", "any", "one", "color") {
		return true
	}
	if len(body) == 8 &&
		effectWordsAt(body, 0, "an", "amount", "of", "mana", "of", "any", "one", "color") {
		return true
	}
	return false
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

// recognizeSourceCounterCountMana types an "Add <mana> for each <kind> counter
// on <this permanent>" add-mana body (Everflowing Chalice) whose produced amount
// scales with the number of counters of one kind on the source permanent.
// parseEffectAmount types the trailing "for each ... counter on this artifact"
// suffix as a source-counter-count dynamic amount, but the leading mana symbol is
// left unparsed because parseEffectMana rejects the trailing count clause. This
// re-parses just the symbols before the count phrase and records the produced
// color so the lowerer can emit one mana per counter. It fires only for a single
// fixed produced color over a recognized counter kind, so choice and any-color
// outputs stay fail-closed.
func recognizeSourceCounterCountMana(effect *EffectSyntax) bool {
	if effect.Kind != EffectAddMana ||
		effect.Amount.DynamicKind != EffectDynamicAmountSourceCounterCount ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormForEach ||
		effect.Amount.Multiplier < 1 ||
		!effect.Amount.CounterKind.Valid() {
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

// recognizeChosenColorCountMana types the add-mana body "an amount of mana of
// that color equal to <dynamic count>" (Three Tree City: "...equal to the number
// of creatures you control of the chosen type."). The trailing count phrase is
// already typed onto effect.Amount as a dynamic battlefield count by
// parseEffectAmount; the leading "an amount of mana of that color" body is left
// unrecognized by parseEffectMana. This credits the chosen-color output when the
// body matches exactly and the amount is a supported battlefield (non-zone)
// count, so the lowerer can produce one mana of the chosen color per counted
// permanent. It fails closed for a card-zone count or a missing amount.
func recognizeChosenColorCountMana(effect *EffectSyntax) bool {
	if effect.Kind != EffectAddMana ||
		effect.Amount.DynamicKind != EffectDynamicAmountCount ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormEqual ||
		effect.Amount.Multiplier < 1 ||
		effect.Amount.Selection == nil ||
		effect.Amount.Selection.Zone != zone.None {
		return false
	}
	body := manaBodyBeforeAmount(effect)
	if len(body) != 7 ||
		!effectWordsAt(body, 0, "an", "amount", "of", "mana", "of", "that", "color") {
		return false
	}
	effect.Mana = EffectManaSyntax{Span: shared.SpanOf(body), ChosenColor: true, ChosenColorDynamic: true}
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

// manaBodyAfterVerb returns the add-mana body tokens that follow the verb,
// dropping a trailing sentence period.
func manaBodyAfterVerb(effect *EffectSyntax) []shared.Token {
	verbEnd := effect.VerbSpan.End.Offset
	var body []shared.Token
	for _, token := range effect.Tokens {
		if token.Span.Start.Offset >= verbEnd {
			body = append(body, token)
		}
	}
	for len(body) > 0 && body[len(body)-1].Kind == shared.Period {
		body = body[:len(body)-1]
	}
	return body
}

// recognizeColorsAmongControlledMana recognizes the add-mana body "one mana of
// any color among <permanents> you control" (Mox Amber, Plaza of Heroes), whose
// produced color is chosen at resolution from the union of colors of the
// permanents the controller controls matching the filter. The filter after
// "among" is parsed by the shared selection parser so it stays generic over the
// permanent group (legendary creatures and planeswalkers, legendary permanents,
// and so on). It fires only for a "you control" battlefield group carrying a
// type, supertype, subtype, or color filter, so an empty wildcard or a non-
// battlefield wording stays fail-closed.
func recognizeColorsAmongControlledMana(effect *EffectSyntax, atoms Atoms) bool {
	if effect.Kind != EffectAddMana ||
		effect.Mana.AnyColor || effect.Mana.ColorsKnown ||
		effect.Mana.CommanderIdentity || effect.Mana.LandsProduce ||
		effect.Mana.LinkedExileColors || effect.Mana.FilterPair ||
		len(effect.Mana.Symbols) != 0 {
		return false
	}
	body := manaBodyAfterVerb(effect)
	if len(body) <= 6 || !effectWordsAt(body, 0, "one", "mana", "of", "any", "color", "among") {
		return false
	}
	selection := parseSelection(body[6:], atoms)
	if selection.Controller != SelectionControllerYou ||
		selection.Zone != zone.None ||
		!colorsAmongSelectionSupported(selection) {
		return false
	}
	clone := selection
	effect.Mana = EffectManaSyntax{
		Span:                  shared.SpanOf(body),
		ColorsAmongControlled: true,
		ColorsAmongSelection:  &clone,
	}
	return true
}

// recognizeEachColorAmongControlledMana recognizes the add-mana body "For each
// color among <permanents> you control, add one mana of that color" (Bloom
// Tender), which produces one mana of each distinct color found among the
// permanents the controller controls matching the filter. The "for each color
// among <group>" prefix precedes the "add" verb; the body after the verb is
// "one mana of that color". The group after "among" is parsed by the shared
// selection parser so it stays generic over the permanent group, and a bare
// "permanents you control" is accepted because the whole controlled board
// legitimately contributes its colors. It fires only for a "you control"
// battlefield group, so a foreign controller or a non-battlefield wording stays
// fail-closed.
func recognizeEachColorAmongControlledMana(effect *EffectSyntax, atoms Atoms) bool {
	if effect.Kind != EffectAddMana ||
		effect.Mana.AnyColor || effect.Mana.ColorsKnown ||
		effect.Mana.ChosenColor || effect.Mana.CommanderIdentity ||
		effect.Mana.LandsProduce || effect.Mana.LinkedExileColors ||
		effect.Mana.FilterPair || effect.Mana.ColorsAmongControlled ||
		effect.Amount.DynamicKind != "" ||
		len(effect.Mana.Symbols) != 0 {
		return false
	}
	body := manaBodyAfterVerb(effect)
	if len(body) != 5 || !effectWordsAt(body, 0, "one", "mana", "of", "that", "color") {
		return false
	}
	prefix := manaPrefixBeforeVerb(effect)
	for len(prefix) > 0 && prefix[len(prefix)-1].Kind == shared.Comma {
		prefix = prefix[:len(prefix)-1]
	}
	if len(prefix) <= 4 || !effectWordsAt(prefix, 0, "for", "each", "color", "among") {
		return false
	}
	selection := parseSelection(prefix[4:], atoms)
	if selection.Controller != SelectionControllerYou ||
		selection.Zone != zone.None {
		return false
	}
	// Accept either a narrowed group whose predicate the selection parser
	// captures (the colorsAmongControlled facets) or the exact literal bare
	// "permanents you control" group. Any other prefix (e.g. "monocolored
	// permanents you control", whose qualifier the selection parser drops
	// silently) fails closed so it cannot lower to a mislabeled ability.
	bareControlled := len(prefix) == 7 && effectWordsAt(prefix, 4, "permanents", "you", "control")
	if !colorsAmongSelectionSupported(selection) && !bareControlled {
		return false
	}
	clone := selection
	effect.Mana = EffectManaSyntax{
		Span:                     shared.SpanOf(effect.Tokens),
		EachColorAmongControlled: true,
		ColorsAmongSelection:     &clone,
	}
	return true
}

// manaPrefixBeforeVerb returns the effect tokens that precede the add-mana verb,
// such as a "For each color among <group>," distributive prefix.
func manaPrefixBeforeVerb(effect *EffectSyntax) []shared.Token {
	verbStart := effect.VerbSpan.Start.Offset
	var prefix []shared.Token
	for _, token := range effect.Tokens {
		if token.Span.End.Offset <= verbStart {
			prefix = append(prefix, token)
		}
	}
	return prefix
}

// filter carries an exact, supported permanent predicate. It requires a type,
// supertype, subtype, or color filter (so a bare "permanents you control" with
// no narrowing predicate fails closed) and rejects qualifiers the executable
// backend cannot represent for this body (numeric, combat, tapped, "another",
// or excluded-type/keyword qualifiers).
func colorsAmongSelectionSupported(selection SelectionSyntax) bool {
	if selection.All || selection.Another || selection.Other ||
		selection.Attacking || selection.Blocking ||
		selection.Tapped || selection.Untapped ||
		selection.MatchManaValue || selection.MatchPower || selection.MatchToughness ||
		selection.Keyword != KeywordUnknown || selection.ExcludedKeyword != KeywordUnknown ||
		len(selection.ExcludedTypes) != 0 || len(selection.ExcludedSupertypes) != 0 ||
		len(selection.ExcludedColors) != 0 || len(selection.Alternatives) != 0 {
		return false
	}
	return len(selection.RequiredTypesAny) != 0 ||
		len(selection.Supertypes) != 0 ||
		len(selection.SubtypesAny) != 0 ||
		len(selection.ColorsAny) != 0 ||
		selection.Colorless || selection.Multicolored ||
		selectionKindNarrowsPermanent(selection.Kind)
}

// selectionKindNarrowsPermanent reports whether a selection Kind names a concrete
// permanent card type (so "creatures you control" narrows) rather than the
// catch-all permanent/any kinds (so "permanents you control" alone does not).
func selectionKindNarrowsPermanent(kind SelectionKind) bool {
	switch kind {
	case SelectionArtifact, SelectionCreature, SelectionEnchantment,
		SelectionLand, SelectionPlaneswalker:
		return true
	default:
		return false
	}
}

func parseHandDiscard(effect *EffectSyntax) HandDiscardSyntax {
	if effect.Kind != EffectDiscard ||
		effect.Context != EffectContextController ||
		!effect.Amount.Known || effect.Amount.Value < 1 ||
		len(effect.Targets) != 0 ||
		len(effect.References) != 0 {
		return HandDiscardSyntax{}
	}
	if exactCardCountEffectSyntax(effect, "Discard", "discards", false) {
		return HandDiscardSyntax{Present: true}
	}
	if exactControllerRandomDiscardSyntax(effect) {
		return HandDiscardSyntax{Present: true, AtRandom: true}
	}
	return HandDiscardSyntax{}
}

// exactControllerRandomDiscardSyntax reconstructs the canonical "discard <N>
// card(s) at random" wording for a controller-context discard of a known fixed
// count. The "at random" suffix marks a random discard, distinguishing it from
// the player-choice discard exactCardCountEffectSyntax recognizes.
func exactControllerRandomDiscardSyntax(effect *EffectSyntax) bool {
	if !effect.Amount.Known || effect.Amount.Value < 1 || effect.Amount.RangeKnown ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormNone {
		return false
	}
	noun := "cards"
	if effect.Amount.Value == 1 {
		noun = "card"
	}
	text := exactEffectClauseText(effect)
	amountText := effectAmountSourceText(effect)
	for _, prefix := range []string{"Discard", "You discard"} {
		if strings.EqualFold(text, fmt.Sprintf("%s %s %s at random.", prefix, amountText, noun)) {
			return true
		}
	}
	return false
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
	case EffectContextEachOtherPlayer:
		return len(effect.Targets) == 0 &&
			strings.EqualFold(text, "Each other player discards their hand.")
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

// parseMassReanimationExchangeEffect recognizes the symmetric mass-reanimation
// sentence "Each player exiles all <type> cards from their graveyard, then
// sacrifices all <type> they control, then puts all cards they exiled this way
// onto the battlefield." (Living Death, Living End, Scrap Mastery). The leading
// type word is singular ("creature"/"artifact") in the exile clause and plural
// ("creatures"/"artifacts") in the sacrifice clause; both must name the same
// card type. The whole sentence collapses to one EffectMassReanimationExchange
// whose Selection carries the card-type filter (parsed from the "all <type>
// cards from their graveyard" sub-phrase), letting the lowering stay text-blind.
// Any other wording fails closed and flows through the generic effect parser.
func parseMassReanimationExchangeEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	words, ok := massReanimationExchangeWords(tokens)
	if !ok {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:       EffectMassReanimationExchange,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		VerbSpan:   words[2].Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Context:    EffectContextEachPlayer,
		Selection:  parseSelection(words[3:9], atoms),
		Exact:      true,
	}}, true
}

// massReanimationExchangeWords returns the non-punctuation tokens of a sentence
// when they match the mass-reanimation exchange template (see
// parseMassReanimationExchangeEffect), and reports whether they matched. The
// returned slice is indexable by the template positions, so callers read the
// "all <type> cards from their graveyard" sub-phrase as words[3:9].
func massReanimationExchangeWords(tokens []shared.Token) ([]shared.Token, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Word {
			words = append(words, token)
		}
	}
	template := []string{
		"each", "player", "exiles", "all", "", "cards", "from", "their", "graveyard",
		"then", "sacrifices", "all", "", "they", "control",
		"then", "puts", "all", "cards", "they", "exiled", "this", "way", "onto", "the", "battlefield",
	}
	if len(words) != len(template) {
		return nil, false
	}
	for offset, want := range template {
		if want == "" {
			continue
		}
		if !equalWord(words[offset], want) {
			return nil, false
		}
	}
	singular := strings.ToLower(words[4].Text)
	if singular != "creature" && singular != "artifact" {
		return nil, false
	}
	if !equalWord(words[12], singular+"s") {
		return nil, false
	}
	return words, true
}

// parseAdditionalLandPlaysEffect recognizes the controller-scoped grant of one
// or more extra land plays for the turn: "Play an additional land this turn.",
// "You may play an additional land this turn.", and the multi-land "... two
// additional lands ..." / "... up to N additional lands ..." variants. The "you
// may" permission is folded into an unconditional allowance (the player is never
// forced to play the extra land). The static "on each of your turns" form is a
// separate static ability and is not matched here.
// parseCastAsThoughFlashEffect recognizes the controller-scoped, turn-scoped
// timing permission "You may cast spells this turn as though they had flash."
// (Borne Upon a Wind, Emergence Zone). The leading "you may" is a permission,
// not a resolving choice, so the effect is modeled unconditionally (like
// parseAdditionalLandPlaysEffect) rather than as an optional cast effect. Any
// other wording fails closed and flows through the generic effect parser.
func parseCastAsThoughFlashEffect(sentence Sentence, tokens []shared.Token) ([]EffectSyntax, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	if len(words) >= 2 && equalWord(words[0], "you") && equalWord(words[1], "may") {
		words = words[2:]
	}
	if len(words) != 9 || !equalWord(words[0], "cast") {
		return nil, false
	}
	castToken := words[0]
	if !equalWord(words[1], "spells") ||
		!equalWord(words[2], "this") ||
		!equalWord(words[3], "turn") ||
		!equalWord(words[4], "as") ||
		!equalWord(words[5], "though") ||
		!equalWord(words[6], "they") ||
		!equalWord(words[7], "had") ||
		!equalWord(words[8], "flash") {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:       EffectCastAsThoughFlash,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		VerbSpan:   castToken.Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), tokens...),
		Context:    EffectContextController,
		Duration:   EffectDurationThisTurn,
		Exact:      true,
	}}, true
}

// parseCantCastSpellsEffect recognizes the one-shot, turn-scoped player cast
// prohibition "<players> can't cast spells this turn." (Silence: "Your opponents
// can't cast spells this turn."; "Players can't cast spells this turn."). The
// affected players are the controller's opponents ("your opponents", "each
// opponent") or every player ("players"). It is modeled as an unconditional
// turn-scoped restriction reusing the continuous cast-prohibition rule effect.
// Targeted, referenced, defending-player, and spell-type-filtered wordings fail
// closed and flow through the generic effect parser.
func parseCantCastSpellsEffect(sentence Sentence, tokens []shared.Token) ([]EffectSyntax, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	index := 0
	allPlayers := false
	switch {
	case len(words) >= 2 && equalWord(words[0], "your") && equalWord(words[1], "opponents"):
		index = 2
	case len(words) >= 2 && equalWord(words[0], "each") && equalWord(words[1], "opponent"):
		index = 2
	case len(words) >= 1 && equalWord(words[0], "players"):
		allPlayers = true
		index = 1
	default:
		return nil, false
	}
	if index >= len(words) || (!equalWord(words[index], "can't") && !equalWord(words[index], "cannot")) {
		return nil, false
	}
	index++
	rest := []string{"cast", "spells", "this", "turn"}
	if len(words)-index != len(rest) {
		return nil, false
	}
	for offset, want := range rest {
		if !equalWord(words[index+offset], want) {
			return nil, false
		}
	}
	return []EffectSyntax{{
		Kind:                     EffectCantCastSpells,
		Span:                     sentence.Span,
		ClauseSpan:               sentence.Span,
		VerbSpan:                 words[index].Span,
		Text:                     sentence.Text,
		Tokens:                   append([]shared.Token(nil), tokens...),
		Context:                  EffectContextController,
		Duration:                 EffectDurationThisTurn,
		CantCastSpellsAllPlayers: allPlayers,
		Exact:                    true,
	}}, true
}

// parseGroupMustAttackEffect recognizes the one-shot, turn-scoped forced-attack
// effect "<group> attack this turn if able." (Bident of Thassa: "Creatures your
// opponents control attack this turn if able."; "Creatures you control attack
// this turn if able."; "All creatures attack this turn if able."). The affected
// creature group is recorded in StaticSubject so lowering can scope the
// continuous must-attack rule effect by controller. Any other subject,
// duration, or trailing clause fails closed and flows through the generic effect
// parser.
func parseGroupMustAttackEffect(sentence Sentence, tokens []shared.Token) ([]EffectSyntax, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	var subject EffectStaticSubjectKind
	index := 0
	switch {
	case len(words) >= 4 && equalWord(words[0], "creatures") &&
		equalWord(words[1], "your") && equalWord(words[2], "opponents") && equalWord(words[3], "control"):
		subject = EffectStaticSubjectOpponentControlledCreatures
		index = 4
	case len(words) >= 3 && equalWord(words[0], "creatures") &&
		equalWord(words[1], "you") && equalWord(words[2], "control"):
		subject = EffectStaticSubjectControlledCreatures
		index = 3
	case len(words) >= 2 && equalWord(words[0], "all") && equalWord(words[1], "creatures"):
		subject = EffectStaticSubjectAllCreatures
		index = 2
	default:
		return nil, false
	}
	rest := []string{"attack", "this", "turn", "if", "able"}
	if len(words)-index != len(rest) {
		return nil, false
	}
	for offset, want := range rest {
		if !equalWord(words[index+offset], want) {
			return nil, false
		}
	}
	return []EffectSyntax{{
		Kind:          EffectMustAttack,
		Span:          sentence.Span,
		ClauseSpan:    sentence.Span,
		VerbSpan:      words[index].Span,
		Text:          sentence.Text,
		Tokens:        append([]shared.Token(nil), tokens...),
		Context:       EffectContextController,
		Duration:      EffectDurationThisTurn,
		StaticSubject: EffectStaticSubjectSyntax{Kind: subject, Span: shared.SpanOf(words[:index])},
		Exact:         true,
	}}, true
}

// resolving buff "The next spell you cast this turn can't be countered."
// (Mistrise Village) and the all-spells form "Spells you cast this turn can't be
// countered." (Domri, Anarch of Bolas). The leading "The next" marks the
// single-next-spell variant; a bare "Spells" marks the every-spell-this-turn
// variant. The buff applies to the controller's own spells, so any other
// subject, a type filter, a negation, or extra wording fails closed and flows
// through the generic effect parser.
func parseSpellsCantBeCounteredEffect(sentence Sentence, tokens []shared.Token) ([]EffectSyntax, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	index := 0
	nextOnly := false
	switch {
	case len(words) >= 3 && equalWord(words[0], "the") && equalWord(words[1], "next") && equalWord(words[2], "spell"):
		nextOnly = true
		index = 3
	case len(words) >= 1 && equalWord(words[0], "spells"):
		index = 1
	default:
		return nil, false
	}
	rest := []string{"you", "cast", "this", "turn"}
	if index+len(rest) > len(words) {
		return nil, false
	}
	for offset, want := range rest {
		if !equalWord(words[index+offset], want) {
			return nil, false
		}
	}
	castToken := words[index+1]
	index += len(rest)
	if index >= len(words) || (!equalWord(words[index], "can't") && !equalWord(words[index], "cannot")) {
		return nil, false
	}
	index++
	tail := []string{"be", "countered"}
	if len(words)-index != len(tail) {
		return nil, false
	}
	for offset, want := range tail {
		if !equalWord(words[index+offset], want) {
			return nil, false
		}
	}
	return []EffectSyntax{{
		Kind:                          EffectSpellsCantBeCountered,
		Span:                          sentence.Span,
		ClauseSpan:                    sentence.Span,
		VerbSpan:                      castToken.Span,
		Text:                          sentence.Text,
		Tokens:                        append([]shared.Token(nil), tokens...),
		Context:                       EffectContextController,
		Duration:                      EffectDurationThisTurn,
		SpellsCantBeCounteredNextOnly: nextOnly,
		Exact:                         true,
	}}, true
}

// parsePreventCombatDamageEffect recognizes the one-shot, turn-scoped combat
// damage prevention shield "Prevent all combat damage that would be dealt to
// [and dealt by] <object> this turn." (Maze of Ith, Goblin Snowman, Moonlight
// Geist), where <object> is a back-reference ("that creature", "this creature",
// "it") to a prior target or the source. PreventDamageTo/PreventDamageBy record
// the prevented directions. Wordings without "this turn" (continuous static
// prevention), with a player or group recipient, or with an unrecognized object
// fail closed and flow through the generic effect parser.
func parsePreventCombatDamageEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	words := make([]shared.Token, 0, len(tokens))
	for _, token := range tokens {
		if token.Kind == shared.Period {
			continue
		}
		words = append(words, token)
	}
	prefix := []string{"prevent", "all", "combat", "damage", "that", "would", "be", "dealt"}
	if len(words) < len(prefix) {
		return nil, false
	}
	for i, want := range prefix {
		if !equalWord(words[i], want) {
			return nil, false
		}
	}
	idx := len(prefix)
	preventTo, preventBy := false, false
	switch {
	case idx+3 < len(words) && equalWord(words[idx], "to") &&
		equalWord(words[idx+1], "and") && equalWord(words[idx+2], "dealt") && equalWord(words[idx+3], "by"):
		preventTo, preventBy = true, true
		idx += 4
	case idx+3 < len(words) && equalWord(words[idx], "by") &&
		equalWord(words[idx+1], "and") && equalWord(words[idx+2], "dealt") && equalWord(words[idx+3], "to"):
		preventTo, preventBy = true, true
		idx += 4
	case idx < len(words) && equalWord(words[idx], "to"):
		preventTo = true
		idx++
	case idx < len(words) && equalWord(words[idx], "by"):
		preventBy = true
		idx++
	default:
		return nil, false
	}
	objectStart := idx
	switch {
	case idx+1 < len(words) && (equalWord(words[idx], "that") || equalWord(words[idx], "this")) && equalWord(words[idx+1], "creature"):
		idx += 2
	case idx < len(words) && equalWord(words[idx], "it"):
		idx++
	default:
		return nil, false
	}
	objectSpan := shared.SpanOf(words[objectStart:idx])
	if idx+2 != len(words) || !equalWord(words[idx], "this") || !equalWord(words[idx+1], "turn") {
		return nil, false
	}
	references := referencesInSpan(atoms, objectSpan)
	if len(references) != 1 {
		return nil, false
	}
	return []EffectSyntax{{
		Kind:            EffectPreventDamage,
		Span:            sentence.Span,
		ClauseSpan:      sentence.Span,
		VerbSpan:        words[0].Span,
		Text:            sentence.Text,
		Tokens:          append([]shared.Token(nil), tokens...),
		Context:         EffectContextController,
		Duration:        EffectDurationThisTurn,
		PreventDamageTo: preventTo,
		PreventDamageBy: preventBy,
		References:      references,
		Exact:           true,
	}}, true
}

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
func parseDamageRecipientPair(kind EffectKind, clause []shared.Token, amount EffectAmountSyntax, atoms Atoms) []SelectionSyntax {
	if kind != EffectDealDamage {
		return nil
	}
	recipient, ok := damageRecipientTokens(clause)
	if !ok {
		recipient, ok = damageRecipientTokensAfterAmount(clause, amount)
		if !ok {
			return nil
		}
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

// damageRecipientTokensAfterAmount returns the recipient tokens of a deal-damage
// clause whose amount is a dynamic "equal to ..." phrase ("deals damage equal to
// its power to each other creature and each opponent."): everything after the
// first "to" that follows the amount span, with a trailing period removed. The
// plain damageRecipientTokens path fails for this wording because "damage" is
// followed by "equal" rather than "to", so the recipient is read from the "to"
// that introduces it after the amount. It fails closed when the amount is not a
// dynamic phrase or no recipient "to" follows it.
func damageRecipientTokensAfterAmount(clause []shared.Token, amount EffectAmountSyntax) ([]shared.Token, bool) {
	if amount.DynamicKind == EffectDynamicAmountNone {
		return nil, false
	}
	for i := 0; i+1 < len(clause); i++ {
		if clause[i].Span.Start.Offset < amount.Span.End.Offset {
			continue
		}
		if !equalWord(clause[i], "to") {
			continue
		}
		recipient := clause[i+1:]
		if len(recipient) > 0 && recipient[len(recipient)-1].Kind == shared.Period {
			recipient = recipient[:len(recipient)-1]
		}
		if len(recipient) == 0 {
			return nil, false
		}
		return recipient, true
	}
	return nil, false
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
// eachSourceDamageSyntax recognizes an "each <group> deals N damage to its
// controller/owner" effect, where every member of the subject group is the
// damage source dealing to the player who controls (or owns) it ("Each creature
// deals 1 damage to its controller."). It returns the parsed source-group
// selection and the recipient role (controller or owner). It fails closed
// (empty selection, None role) for every other shape: a non-damage effect, a
// subject that does not begin with "each" or does not parse to a recognized
// group, or a recipient that is not the bare "its controller"/"its owner".
func eachSourceDamageSyntax(kind EffectKind, subject, clause []shared.Token, atoms Atoms) (SelectionSyntax, DamageRecipientReferenceKind) {
	if kind != EffectDealDamage || len(subject) == 0 || !equalWord(subject[0], "each") {
		return SelectionSyntax{}, DamageRecipientReferenceNone
	}
	recipient, ok := damageRecipientTokens(clause)
	if !ok || len(recipient) != 2 || !equalWord(recipient[0], "its") {
		return SelectionSyntax{}, DamageRecipientReferenceNone
	}
	var role DamageRecipientReferenceKind
	switch {
	case equalWord(recipient[1], "controller"):
		role = DamageRecipientReferenceController
	case equalWord(recipient[1], "owner"):
		role = DamageRecipientReferenceOwner
	default:
		return SelectionSyntax{}, DamageRecipientReferenceNone
	}
	selection := parseSelection(subject, atoms)
	if selection.Kind == SelectionUnknown {
		return SelectionSyntax{}, DamageRecipientReferenceNone
	}
	return selection, role
}

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
	return effect.Mana.AnyColor || effect.Mana.CommanderIdentity || effect.Mana.LandsProduce || effect.Mana.FilterPair || effect.Mana.ColorsAmongControlled || len(effect.Mana.Symbols) != 0
}

func legacyEffectCount(tokens []shared.Token, atoms Atoms) int {
	if _, ok := massReanimationExchangeWords(tokens); ok {
		return 1
	}
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
	case kind == EffectCast && pastCastCountPhraseAt(tokens, index):
		return EffectUnknown
	case kind == EffectCast && castDuringMainPhaseConditionAt(tokens, index):
		return EffectUnknown
	case kind == EffectCast && castSpellsFromLibraryTopAt(tokens, index):
		return EffectUnknown
	case kind == EffectCounter && !counterVerbAt(tokens, index):
		return EffectUnknown
	case kind == EffectCopyStackObject && !copyVerbAt(tokens, index):
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

// groupEntersTappedPermanentType maps a plural permanent-type noun used as the
// subject of a static "<permanents> enter tapped" replacement to its runtime
// card type. It reports ok=false for any word that is not a recognized
// permanent-type plural so the caller's type list fails closed.
func groupEntersTappedPermanentType(word string) (types.Card, bool) {
	switch strings.ToLower(word) {
	case "creatures":
		return types.Creature, true
	case "lands":
		return types.Land, true
	case "artifacts":
		return types.Artifact, true
	case "enchantments":
		return types.Enchantment, true
	case "planeswalkers":
		return types.Planeswalker, true
	default:
		return "", false
	}
}

// parseEntersAsCopyEffect recognizes a self enters-the-battlefield replacement
// that has the permanent enter as a copy of another permanent chosen as it
// enters ("You may have this creature enter the battlefield as a copy of any
// creature on the battlefield.", Clone; CR 706). The copied-permanent filter is
// the noun phrase after "as a copy of", up to an optional ", except <rider>"
// clause. Only the "isn't legendary" and "is an <type> in addition to its other
// types" copiable riders are recognized; any other rider fails closed so the
// card stays unsupported.
func parseEntersAsCopyEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	body := semanticEffectTokens(tokens)
	if len(body) < 5 || body[len(body)-1].Kind != shared.Period {
		return nil, false
	}
	// An "As this <permanent> enters," replacement prefix frames the temporary
	// "become a copy ... until end of turn" form (Cursed Mirror). Strip it so the
	// shared subject scan below sees the "you may have it ..." body, and record
	// that the enter verb was consumed by the prefix.
	viaAsEnters := false
	if afterPrefix, ok := entersAsCopyAsEntersPrefix(body); ok {
		body = body[afterPrefix:]
		viaAsEnters = true
	}
	words := normalizedWords(body)
	// Only a self enters-as-copy is supported: "You may have this <permanent>
	// enter ..." or "This <permanent> enters ...". Group forms such as
	// "Creatures you control enter as a copy of this creature." (Essence of the
	// Wild) have a different subject and fail closed.
	selfSubject := len(words) >= 1 && words[0] == "this" ||
		len(words) >= 3 && words[0] == "you" && words[1] == "may" && words[2] == "have"
	if !selfSubject {
		return nil, false
	}
	copyIndex := -1
	for i := 2; i+1 < len(body); i++ {
		if equalWord(body[i], "copy") && equalWord(body[i-1], "a") && equalWord(body[i+1], "of") &&
			(equalWord(body[i-2], "as") || equalWord(body[i-2], "become")) {
			copyIndex = i
			break
		}
	}
	if copyIndex < 0 {
		return nil, false
	}
	enters := viaAsEnters
	for i := 0; i < copyIndex; i++ {
		if equalWord(body[i], "tapped") {
			// "enter tapped as a copy" (Vesuva) also overrides the enters-tapped
			// state, which this replacement does not model; fail closed.
			return nil, false
		}
		if equalWord(body[i], "enter") || equalWord(body[i], "enters") {
			enters = true
		}
	}
	if !enters {
		return nil, false
	}
	filterStart := copyIndex + 2
	filterEnd := len(body) - 1
	var notLegendary bool
	var addTypes []types.Card
	var addKeywords []KeywordKind
	var conditionalCounters []EntersAsCopyConditionalCounter
	if exceptIndex := entersAsCopyExceptIndex(body, filterStart); exceptIndex >= 0 {
		riders, ok := parseEntersAsCopyRider(body[exceptIndex+1:len(body)-1], atoms)
		if !ok {
			return nil, false
		}
		notLegendary = riders.notLegendary
		addTypes = riders.addTypes
		addKeywords = riders.addKeywords
		conditionalCounters = riders.conditionalCounters
		filterEnd = exceptIndex
	}
	filter := body[filterStart:filterEnd]
	for len(filter) > 0 && filter[len(filter)-1].Kind == shared.Comma {
		filter = filter[:len(filter)-1]
	}
	// "become a copy ... until end of turn" makes the copy a temporary effect
	// (Cursed Mirror); strip the trailing duration phrase and record it so
	// lowering scopes the copy effect to end of turn.
	untilEndOfTurn := false
	if trimmed, ok := trimTrailingUntilEndOfTurn(filter); ok {
		filter = trimmed
		untilEndOfTurn = true
	}
	// Only battlefield permanents can be copied by this runtime, so require an
	// explicit battlefield or "you control" scope; graveyard/hand sources such
	// as Body Double ("any creature card in a graveyard") fail closed.
	if !entersAsCopyFilterOnBattlefield(filter) {
		return nil, false
	}
	filter = trimTrailingZonePhrase(filter)
	if len(filter) == 0 {
		return nil, false
	}
	optional := words[0] == "you"
	effect := EffectSyntax{
		Kind:                     EffectEnterAsCopy,
		Context:                  EffectContextController,
		Span:                     sentence.Span,
		ClauseSpan:               sentence.Span,
		Text:                     sentence.Text,
		Tokens:                   append([]shared.Token(nil), body...),
		Selection:                parseSelection(filter, atoms),
		EntersAsCopy:             true,
		EntersAsCopyOptional:     optional,
		EntersAsCopyNotLegendary: notLegendary,
		EntersAsCopyAddTypes:     addTypes,

		EntersAsCopyConditionalCounters: conditionalCounters,
		EntersAsCopyUntilEndOfTurn:      untilEndOfTurn,
		EntersAsCopyAddKeywords:         addKeywords,
	}
	return []EffectSyntax{effect}, true
}

// parseBecomeCopyEffect recognizes an activated/resolving copy effect that has
// the source permanent become a copy of a targeted permanent ("This land
// becomes a copy of target land, except it has this ability.", Thespian's Stage;
// "This artifact becomes a copy of target ... until end of turn.", Mirage
// Mirror; CR 706). The copied-permanent target is left as the ordinary "target"
// selector for the target machinery to extract; only the "until end of turn"
// duration and the "except it has this ability" / "except it has <keyword>"
// copiable riders are recognized, and any other rider fails closed.
func parseBecomeCopyEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	body := semanticEffectTokens(tokens)
	if len(body) < 6 || body[len(body)-1].Kind != shared.Period {
		return nil, false
	}
	words := normalizedWords(body)
	// The subject must be the source permanent itself ("This <permanent> becomes
	// a copy of ..."). Group or other-subject copies fail closed.
	if len(words) < 2 || words[0] != "this" {
		return nil, false
	}
	copyIndex := -1
	for i := 3; i+2 < len(body); i++ {
		if equalWord(body[i], "copy") && equalWord(body[i-1], "a") && equalWord(body[i-2], "becomes") &&
			equalWord(body[i+1], "of") && equalWord(body[i+2], "target") {
			copyIndex = i
			break
		}
	}
	if copyIndex < 0 {
		return nil, false
	}
	rest := body[copyIndex+3 : len(body)-1]
	var untilEndOfTurn, retainAbility bool
	var addKeywords []KeywordKind
	if exceptIndex := entersAsCopyExceptIndex(rest, 0); exceptIndex >= 0 {
		retain, keywords, ok := parseBecomeCopyRider(rest[exceptIndex+1:], atoms)
		if !ok {
			return nil, false
		}
		retainAbility = retain
		addKeywords = keywords
		rest = rest[:exceptIndex]
	}
	for len(rest) > 0 && rest[len(rest)-1].Kind == shared.Comma {
		rest = rest[:len(rest)-1]
	}
	if trimmed, ok := becomeCopyTrimUntilEndOfTurn(rest); ok {
		rest = trimmed
		untilEndOfTurn = true
	}
	// A target selector must remain ("land", "artifact, creature, ...").
	if len(rest) == 0 {
		return nil, false
	}
	effect := EffectSyntax{
		Kind:                         EffectBecomeCopy,
		Context:                      EffectContextController,
		Span:                         sentence.Span,
		ClauseSpan:                   sentence.Span,
		Text:                         sentence.Text,
		Tokens:                       append([]shared.Token(nil), body...),
		BecomeCopyUntilEndOfTurn:     untilEndOfTurn,
		BecomeCopyRetainsThisAbility: retainAbility,
		BecomeCopyAddKeywords:        addKeywords,
	}
	return []EffectSyntax{effect}, true
}

// becomeCopyTrimUntilEndOfTurn removes a trailing "until end of turn" phrase from
// rest, reporting whether it was present.
func becomeCopyTrimUntilEndOfTurn(rest []shared.Token) ([]shared.Token, bool) {
	if len(rest) < 4 {
		return rest, false
	}
	tail := rest[len(rest)-4:]
	if equalWord(tail[0], "until") && equalWord(tail[1], "end") &&
		equalWord(tail[2], "of") && equalWord(tail[3], "turn") {
		return rest[:len(rest)-4], true
	}
	return rest, false
}

// parseBecomeCopyRider parses the copiable riders of a become-a-copy "except
// <rider>" clause. It recognizes "it has this ability" (RetainsThisAbility) and
// "it has <keyword>" keyword riders; any other rider fails closed.
func parseBecomeCopyRider(clause []shared.Token, atoms Atoms) (bool, []KeywordKind, bool) {
	words := normalizedWords(clause)
	// Drop a leading "it has" / "it's" / "it is" subject.
	switch {
	case len(words) >= 2 && words[0] == "it" && words[1] == "has":
		clause = clause[2:]
		words = words[2:]
	case len(words) >= 2 && words[0] == "it" && words[1] == "is":
		clause = clause[2:]
		words = words[2:]
	default:
		return false, nil, false
	}
	if len(words) == 2 && words[0] == "this" && words[1] == "ability" {
		return true, nil, true
	}
	keywords := scanKeywords(clause, atoms)
	if len(keywords) == 0 {
		return false, nil, false
	}
	kinds := make([]KeywordKind, 0, len(keywords))
	for _, keyword := range keywords {
		kinds = append(kinds, keyword.Kind)
	}
	return false, kinds, true
}

// entersAsCopyAsEntersPrefix reports whether body begins with an "As this
// <permanent> enters," replacement prefix and, when it does, returns the index
// of the first token after the introducing comma. The prefix must reach its
// enter verb before any comma so unrelated "As ..." clauses fail closed.
func entersAsCopyAsEntersPrefix(body []shared.Token) (int, bool) {
	if len(body) < 4 || !equalWord(body[0], "as") || !equalWord(body[1], "this") {
		return 0, false
	}
	for i := 2; i < len(body); i++ {
		if body[i].Kind == shared.Comma {
			return 0, false
		}
		if equalWord(body[i], "enter") || equalWord(body[i], "enters") {
			if i+1 < len(body) && body[i+1].Kind == shared.Comma {
				return i + 2, true
			}
			return 0, false
		}
	}
	return 0, false
}

// trimTrailingUntilEndOfTurn removes a trailing "until end of turn" phrase from a
// copied-permanent filter, reporting whether it was present. The phrase marks the
// temporary "become a copy ... until end of turn" copy duration (Cursed Mirror).
func trimTrailingUntilEndOfTurn(filter []shared.Token) ([]shared.Token, bool) {
	if len(filter) < 4 {
		return filter, false
	}
	tail := filter[len(filter)-4:]
	if equalWord(tail[0], "until") && equalWord(tail[1], "end") &&
		equalWord(tail[2], "of") && equalWord(tail[3], "turn") {
		return filter[:len(filter)-4], true
	}
	return filter, false
}

// entersAsCopyExceptIndex finds the index of the "except" word that introduces a
// copiable rider in an enters-as-copy clause, searching from start. It returns
// -1 when no rider is present.
func entersAsCopyExceptIndex(body []shared.Token, start int) int {
	for i := start; i < len(body); i++ {
		if equalWord(body[i], "except") {
			return i
		}
	}
	return -1
}

// entersAsCopyRiders collects the recognized copiable riders parsed from an
// enters-as-copy "except <rider>" clause.
type entersAsCopyRiders struct {
	addTypes            []types.Card
	addKeywords         []KeywordKind
	notLegendary        bool
	conditionalCounters []EntersAsCopyConditionalCounter
}

// parseEntersAsCopyRider parses the recognized copiable riders of an
// enters-as-copy clause: "it isn't legendary" / "it's not legendary" sets the
// not-legendary flag, "it's an <type> in addition to its other types" adds the
// named card type, and "it enters with an additional <kind> counter on it if
// it's a <type>" adds a conditional copiable counter (Spark Double). Riders
// joined by commas or "and" are supported. Each clause must match a recognized
// template exactly; any other wording fails closed.
func parseEntersAsCopyRider(rider []shared.Token, atoms Atoms) (entersAsCopyRiders, bool) {
	clauses := splitEntersAsCopyRiderClauses(rider)
	if len(clauses) == 0 {
		return entersAsCopyRiders{}, false
	}
	var riders entersAsCopyRiders
	for _, clause := range clauses {
		words := normalizedWords(clause)
		if entersAsCopyNotLegendaryClause(words) {
			riders.notLegendary = true
			continue
		}
		if placement, ok := entersAsCopyConditionalCounterClause(clause); ok {
			riders.conditionalCounters = append(riders.conditionalCounters, placement)
			continue
		}
		if keyword, ok := entersAsCopyAddKeywordClause(clause, atoms); ok {
			riders.addKeywords = append(riders.addKeywords, keyword)
			continue
		}
		cardType, typeOK := entersAsCopyAddTypeClause(words)
		if !typeOK {
			return entersAsCopyRiders{}, false
		}
		riders.addTypes = append(riders.addTypes, cardType)
	}
	return riders, true
}

// entersAsCopyAddKeywordClause matches the "it has <keyword>" copiable rider on
// an enters-as-copy replacement ("except it has haste", Cursed Mirror) and
// returns the single granted keyword. It fails closed on any wording other than
// exactly "it has <keyword>" with one recognized keyword filling the clause.
func entersAsCopyAddKeywordClause(clause []shared.Token, atoms Atoms) (KeywordKind, bool) {
	if len(clause) < 3 || !equalWord(clause[0], "it") || !equalWord(clause[1], "has") {
		return KeywordUnknown, false
	}
	keywords := scanKeywords(clause[2:], atoms)
	if len(keywords) != 1 {
		return KeywordUnknown, false
	}
	return keywords[0].Kind, true
}

// splitEntersAsCopyRiderClauses splits a copiable-rider token run into individual
// clauses on top-level commas and "and" conjunctions.
func splitEntersAsCopyRiderClauses(rider []shared.Token) [][]shared.Token {
	var clauses [][]shared.Token
	var current []shared.Token
	flush := func() {
		if len(current) > 0 {
			clauses = append(clauses, current)
			current = nil
		}
	}
	for _, tok := range rider {
		if tok.Kind == shared.Comma || equalWord(tok, "and") {
			flush()
			continue
		}
		current = append(current, tok)
	}
	flush()
	return clauses
}

// entersAsCopyNotLegendaryClause reports whether a rider clause is the
// "it isn't legendary" / "it's not legendary" copiable rider.
func entersAsCopyNotLegendaryClause(words []string) bool {
	if len(words) < 2 || len(words) > 4 {
		return false
	}
	if words[len(words)-1] != "legendary" {
		return false
	}
	negation := false
	for _, word := range words[:len(words)-1] {
		switch word {
		case "it", "it's", "is":
		case "isn't", "not", "n't":
			negation = true
		default:
			return false
		}
	}
	return negation
}

// entersAsCopyAddTypeClause matches the "it's an <type> in addition to its other
// types" copiable rider exactly and returns the single added card type. It fails
// closed on any other wording, including subtype additions such as "a Synth
// artifact" that this replacement cannot represent.
func entersAsCopyAddTypeClause(words []string) (types.Card, bool) {
	switch {
	case len(words) == 9 && words[0] == "it's":
		words = words[1:]
	case len(words) == 10 && words[0] == "it" && (words[1] == "is" || words[1] == "'s"):
		words = words[2:]
	default:
		return "", false
	}
	if words[0] != "a" && words[0] != "an" {
		return "", false
	}
	if words[2] != "in" || words[3] != "addition" || words[4] != "to" ||
		words[5] != "its" || words[6] != "other" || words[7] != "types" {
		return "", false
	}
	return entersAsCopyAddTypeWord(words[1])
}

// entersAsCopyAddTypeWord maps a singular card-type word used in an
// enters-as-copy "in addition to its other types" rider to its card type. It
// fails closed on any other word.
func entersAsCopyAddTypeWord(word string) (types.Card, bool) {
	switch strings.ToLower(word) {
	case "artifact":
		return types.Artifact, true
	case "creature":
		return types.Creature, true
	case "enchantment":
		return types.Enchantment, true
	case "land":
		return types.Land, true
	default:
		return "", false
	}
}

// entersAsCopyConditionalCounterClause matches the conditional copiable counter
// rider "it enters with an additional <kind> counter on it if it's a <type>"
// (Spark Double) and returns a single counter placement guarded by the named
// card type. The counter kind is read at the token level (counterNameBefore)
// because symbol kinds such as "+1/+1" are dropped from normalized words. It
// fails closed on any other wording.
func entersAsCopyConditionalCounterClause(clause []shared.Token) (EntersAsCopyConditionalCounter, bool) {
	counterIndex := entersAsCopyWordIndex(clause, "counter", 0)
	if counterIndex < 0 {
		return EntersAsCopyConditionalCounter{}, false
	}
	kind, _, ok := counterNameBefore(clause, counterIndex)
	if !ok {
		return EntersAsCopyConditionalCounter{}, false
	}
	prefix := normalizedWords(clause[:counterIndex])
	if !slices.Contains(prefix, "enters") && !slices.Contains(prefix, "enter") {
		return EntersAsCopyConditionalCounter{}, false
	}
	if !slices.Contains(prefix, "with") {
		return EntersAsCopyConditionalCounter{}, false
	}
	ifIndex := entersAsCopyWordIndex(clause, "if", counterIndex)
	if ifIndex < 0 {
		return EntersAsCopyConditionalCounter{}, false
	}
	cardType, ok := entersAsCopyConditionalTypeTail(normalizedWords(clause[ifIndex:]))
	if !ok {
		return EntersAsCopyConditionalCounter{}, false
	}
	return EntersAsCopyConditionalCounter{Kind: kind, Amount: 1, IfType: cardType}, true
}

// entersAsCopyConditionalTypeTail parses the "if it's a <type>" / "if it is a
// <type>" tail of a conditional copiable counter rider into its card type. It
// fails closed on any other wording.
func entersAsCopyConditionalTypeTail(words []string) (types.Card, bool) {
	switch {
	case len(words) == 4 && words[0] == "if" && words[1] == "it's" && (words[2] == "a" || words[2] == "an"):
		return entersAsCopyConditionalTypeWord(words[3])
	case len(words) == 5 && words[0] == "if" && words[1] == "it" && words[2] == "is" && (words[3] == "a" || words[3] == "an"):
		return entersAsCopyConditionalTypeWord(words[4])
	default:
		return "", false
	}
}

// entersAsCopyConditionalTypeWord maps a singular card-type word used in a
// conditional copiable counter rider's "if it's a <type>" tail to its card
// type, including planeswalker (loyalty rider). It fails closed on any other
// word.
func entersAsCopyConditionalTypeWord(word string) (types.Card, bool) {
	switch strings.ToLower(word) {
	case "artifact":
		return types.Artifact, true
	case "creature":
		return types.Creature, true
	case "enchantment":
		return types.Enchantment, true
	case "land":
		return types.Land, true
	case "planeswalker":
		return types.Planeswalker, true
	default:
		return "", false
	}
}

// entersAsCopyWordIndex returns the index of the first token at or after start
// whose word equals word, or -1 when none matches.
func entersAsCopyWordIndex(tokens []shared.Token, word string, start int) int {
	for i := start; i < len(tokens); i++ {
		if equalWord(tokens[i], word) {
			return i
		}
	}
	return -1
}

// entersAsCopyFilterOnBattlefield reports whether a copy-filter token run is
// scoped to the battlefield, either via a trailing "on the battlefield" phrase
// or a "you control" controller relation. Filters without such a scope (for
// example "any creature card in a graveyard") fail closed because this
// replacement can only copy battlefield permanents.
func entersAsCopyFilterOnBattlefield(filter []shared.Token) bool {
	if len(filter) >= 3 &&
		equalWord(filter[len(filter)-3], "on") &&
		equalWord(filter[len(filter)-2], "the") &&
		equalWord(filter[len(filter)-1], "battlefield") {
		return true
	}
	for i := 0; i+1 < len(filter); i++ {
		if equalWord(filter[i], "you") && equalWord(filter[i+1], "control") {
			return true
		}
	}
	return false
}

// trimTrailingZonePhrase drops a trailing "on the battlefield" zone phrase from
// a copy-filter token run so the filter selects the permanent type rather than a
// battlefield zone, which would make the selector unrepresentable.
func trimTrailingZonePhrase(filter []shared.Token) []shared.Token {
	if len(filter) >= 3 &&
		equalWord(filter[len(filter)-3], "on") &&
		equalWord(filter[len(filter)-2], "the") &&
		equalWord(filter[len(filter)-1], "battlefield") {
		return filter[:len(filter)-3]
	}
	return filter
}

// parseGroupEntersTappedEffect recognizes a static enters-tapped replacement that
// taps a group of OTHER permanents as they enter, such as "Creatures your
// opponents control enter tapped." (Authority of the Consuls), "Artifacts,
// creatures, and lands your opponents control enter the battlefield tapped."
// (Frozen Aether), or the unscoped "Permanents enter tapped." (Kismet family).
// The subject is a list of permanent-type plurals (or the catch-all
// "Permanents"), an optional controller scope, and the plural "enter [the
// battlefield] tapped" verb phrase. It matches the whole sentence exactly, so
// any other wording falls through to the generic effect grammar.
func parseGroupEntersTappedEffect(sentence Sentence, tokens []shared.Token) ([]EffectSyntax, bool) {
	body := semanticEffectTokens(tokens)
	if len(body) < 4 || body[len(body)-1].Kind != shared.Period {
		return nil, false
	}
	words := body[:len(body)-1]
	index := 0
	var cardTypes []types.Card
	if equalWord(words[0], "permanents") {
		index = 1
	} else {
		for index < len(words) {
			cardType, ok := groupEntersTappedPermanentType(words[index].Text)
			if !ok {
				break
			}
			cardTypes = append(cardTypes, cardType)
			index++
			for index < len(words) && (words[index].Kind == shared.Comma || equalWord(words[index], "and")) {
				index++
			}
		}
		if len(cardTypes) == 0 {
			return nil, false
		}
	}
	scope := EntersTappedGroupControllerEach
	switch {
	case index+2 < len(words) && equalWord(words[index], "your") &&
		equalWord(words[index+1], "opponents") && equalWord(words[index+2], "control"):
		scope = EntersTappedGroupControllerOpponents
		index += 3
	case index+2 < len(words) && equalWord(words[index], "an") &&
		equalWord(words[index+1], "opponent") && equalWord(words[index+2], "controls"):
		scope = EntersTappedGroupControllerOpponents
		index += 3
	case index+1 < len(words) && equalWord(words[index], "you") && equalWord(words[index+1], "control"):
		scope = EntersTappedGroupControllerYou
		index += 2
	default:
	}
	if index >= len(words) || !equalWord(words[index], "enter") {
		return nil, false
	}
	index++
	if index+1 < len(words) && equalWord(words[index], "the") && equalWord(words[index+1], "battlefield") {
		index += 2
	}
	if index >= len(words) || !equalWord(words[index], "tapped") {
		return nil, false
	}
	index++
	if index != len(words) {
		return nil, false
	}
	effect := EffectSyntax{
		Kind:                   EffectEnterTapped,
		Context:                EffectContextController,
		Span:                   sentence.Span,
		ClauseSpan:             sentence.Span,
		Text:                   sentence.Text,
		Tokens:                 append([]shared.Token(nil), tokens...),
		EntersTapped:           true,
		EntersTappedGroup:      true,
		EntersTappedGroupScope: scope,
		EntersTappedGroupTypes: cardTypes,
	}
	effect.Exact = exactEffectSyntax(&effect)
	return []EffectSyntax{effect}, true
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

// moveAllCountersClause reports the kind-agnostic "move all counters" form,
// where every counter on the source moves regardless of kind ("Move all
// counters from this permanent onto target creature."). It anchors on the
// literal "all counters" run so the specific-kind form ("a +1/+1 counter")
// keeps MoveCountersAll false and lowers through its named-kind path.
func moveAllCountersClause(clause []shared.Token) bool {
	return effectHasTokenWords(clause, "all", "counters")
}
