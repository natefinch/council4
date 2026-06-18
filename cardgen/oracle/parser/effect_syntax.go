package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

func emitResolvingSyntax(abilities []Ability) {
	for i := range abilities {
		emitSentenceResolvingSyntax(abilities[i].Sentences, abilities[i].Atoms, abilities[i].ActivationRestrictions, abilities[i].TriggerFrequency)
		if abilities[i].Modal == nil {
			continue
		}
		for j := range abilities[i].Modal.Options {
			mode := &abilities[i].Modal.Options[j]
			emitSentenceResolvingSyntax(mode.Sentences, mode.Atoms, nil, nil)
		}
	}
}

func emitSentenceResolvingSyntax(sentences []Sentence, atoms Atoms, restrictions []ActivationRestriction, triggerFrequency *TriggerFrequencyRestriction) {
	legacyEffects := 0
	currentEffects := 0
	unrecognizedSibling := false
	for i := range sentences {
		if sentences[i].StaticRule != nil ||
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
		currentEffects += len(sentences[i].Effects)
		if len(tokens) > 0 && len(sentences[i].Effects) == 0 &&
			len(atoms.KeywordsWithin(tokens)) == 0 && count == 0 &&
			!effectWordsAt(tokens, 0, "activate", "only", "if") {
			unrecognizedSibling = true
		}
	}
	if currentEffects == 1 && unrecognizedSibling {
		for i := range sentences {
			for j := range sentences[i].Effects {
				sentences[i].Effects[j].Exact = false
				sentences[i].Effects[j].HasUnrecognizedSibling = true
			}
		}
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

func parseEffects(sentence Sentence, tokens []shared.Token, atoms Atoms) []EffectSyntax {
	indices := effectIndices(tokens, atoms)
	requiresOrderedLowering := legacyEffectCount(tokens, atoms) > 1
	effects := make([]EffectSyntax, 0, len(indices))
	for effectIndex, tokenIndex := range indices {
		clauseEnd := resolvingClauseEnd(tokens, indices, effectIndex)
		ownershipStart := resolvingClauseStart(tokens, indices, effectIndex)
		ownership := tokens[ownershipStart:clauseEnd]
		clause := tokens[tokenIndex+1 : clauseEnd]
		clause, delayed := cutDelayedTiming(clause)
		power, toughness := parsePTChange(clause)
		counterKind, counterKnown := parseCounterPlacement(clause, atoms)
		span := shared.SpanOf(clause)
		ownershipSpan := shared.SpanOf(ownership)
		toZone := firstZone(atoms, span, ZoneRoleTo)
		if ambiguousZoneChoice(ownership, atoms, span) {
			toZone = zone.None
		}
		staticSubject := parseEffectStaticSubject(ownership, atoms)
		payment := parseEffectPayment(tokens)
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
		tokenPower, tokenToughness, tokenPTKnown := parseTokenPowerToughness(kind, clause)
		amount := parseEffectAmount(kind, clause, atoms)
		if forEach, ok := parseCreateForEachAmount(kind, context, tokenPTKnown, tokens[ownershipStart:tokenIndex], amount, atoms); ok {
			amount = forEach
		}
		effects = append(effects, EffectSyntax{
			Kind:                    kind,
			Context:                 context,
			Connection:              connection,
			ConnectionSpan:          connectionSpan,
			Span:                    sentence.Span,
			VerbSpan:                tokens[tokenIndex].Span,
			ClauseSpan:              ownershipSpan,
			Text:                    sentence.Text,
			Tokens:                  append([]shared.Token(nil), ownership...),
			Duration:                parseEffectDuration(durationTokens, atoms),
			DelayedTiming:           delayed,
			Selection:               parseSelection(clause, atoms),
			DamageRecipientPair:     parseDamageRecipientPair(kind, clause, atoms),
			Amount:                  amount,
			PowerDelta:              power,
			ToughnessDelta:          toughness,
			TokenPower:              tokenPower,
			TokenToughness:          tokenToughness,
			TokenPTKnown:            tokenPTKnown,
			StaticSubject:           staticSubject,
			CounterKind:             counterKind,
			CounterKnown:            counterKnown,
			FromZone:                firstZone(atoms, span, ZoneRoleFrom),
			ToZone:                  toZone,
			Destination:             parseEffectDestination(ownership),
			EntersTapped:            effectWordsAtAny(ownership, "battlefield", "tapped"),
			EntersTappedSelf:        entersTappedSelfSyntax(kind, clause),
			EntersColorChoice:       entersColorChoiceSyntax(kind, clause),
			EntersWithCounters:      entersWithCountersSyntax(kind, clause),
			UnderYourControl:        effectContainsWords(normalizedWords(ownership), "under", "your", "control"),
			CastAsAdventure:         effectContainsWords(normalizedWords(clause), "as", "an", "adventure"),
			Negated:                 effectIsNegated(tokens, tokenIndex),
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
		effects[i].Divided = dividedDamageEffect(&effects[i])
		effects[i].Exact = exactEffectSyntax(&effects[i])
		effects[i].TokenCopyOfTarget = exactCreateCopyTokenEffectSyntax(&effects[i])
		effects[i].Mana.LegacyBodyExact = legacyExactManaBody(&effects[i], sentence)
		if effects[i].Kind == EffectSearch {
			effects[i].UnsupportedDetail = searchUnsupportedDetail(&effects[i])
		}
	}
	return effects
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
	return effect.Mana.AnyColor || effect.Mana.CommanderIdentity || len(effect.Mana.Symbols) != 0
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

// entersColorChoiceSyntax recognizes the exact self entry color-choice clause
// "choose a color ." following an "As this <permanent> enters," verb. The enters
// verb is shared by many entry constructs, so this matches only the exact color
// choice; "choose a color other than <color>" and non-color choices fail closed.
func entersColorChoiceSyntax(kind EffectKind, clause []shared.Token) bool {
	if kind != EffectEnterTapped {
		return false
	}
	body := clause
	if len(body) > 0 && body[0].Kind == shared.Comma {
		body = body[1:]
	}
	return len(body) == 4 &&
		equalWord(body[0], "choose") &&
		equalWord(body[1], "a") &&
		equalWord(body[2], "color") &&
		body[3].Text == "."
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
