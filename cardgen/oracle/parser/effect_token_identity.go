package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// parsePassiveTokenIdentityEffects recognizes the passive-voice identity
// substitution token-creation replacement "If one or more [type] tokens would be
// created under your control, that many <substitute token> are created instead."
// (Divine Visitation: "... that many 4/4 white Angel creature tokens with flying
// and vigilance are created instead."). Unlike the doubling and additive forms,
// the resolving clause replaces each created token with one copy of a fully
// spelled-out substitute token rather than creating more of the original. The
// wording carries no active "create" verb, so it is recognized here and emitted
// as two EffectCreate instructions: the would-create group (carrying the
// optional card-type filter in its selector) and the substitute output marked
// EffectReplacementThatManyIdentity. The matching intervening-if condition is
// recognized separately by recognizeTokenCreationCondition.
func parsePassiveTokenIdentityEffects(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	commaIndex, anyController, ok := matchPassiveTokenIdentity(tokens)
	if !ok {
		return nil, false
	}
	condition := tokens[:commaIndex]
	resolving := tokens[commaIndex+1:]

	// The would-create noun phrase ("one or more [type] tokens") is the condition
	// with its leading "if" and trailing "would be created [under your control]"
	// stripped, mirroring the additive form so the optional card-type filter the
	// lowering reads stays on the group's selector.
	nounPhrase := condition
	if len(nounPhrase) > 0 && equalWord(nounPhrase[0], "if") {
		nounPhrase = nounPhrase[1:]
	}
	if trimmed, stripped := stripTokenSuffix(nounPhrase, "would", "be", "created", "under", "your", "control"); stripped {
		nounPhrase = trimmed
	} else if trimmed, stripped := stripTokenSuffix(nounPhrase, "would", "be", "created"); stripped {
		nounPhrase = trimmed
	}
	createdIndex := commaIndex - 1
	for i := range condition {
		if equalWord(condition[i], "created") {
			createdIndex = i
			break
		}
	}
	createEffect := EffectSyntax{
		Kind:             EffectCreate,
		Context:          EffectContextController,
		Span:             shared.SpanOf(condition),
		VerbSpan:         tokens[createdIndex].Span,
		ClauseSpan:       shared.SpanOf(condition),
		Text:             sentence.Text,
		Tokens:           append([]shared.Token(nil), condition...),
		Selection:        parseSelection(nounPhrase, atoms),
		Amount:           EffectAmountSyntax{Value: 1, Known: true},
		UnderYourControl: !anyController,
	}
	if conjunctiveTypeTarget(createEffect.Selection) {
		createEffect.Selection.ConjunctiveTypes = true
	}

	// The substitute clause ("that many <token characteristics>") is the resolving
	// clause with its leading "that many" and trailing "are created instead ."
	// stripped. It is parsed by the same token-characteristic helpers the active
	// create-verb path uses, so the substitute's power/toughness, subtypes,
	// colors, name, and keywords match the active form exactly. Keywords are read
	// from the full output clause (the "with flying and vigilance" rider), while
	// the selection is parsed from the noun phrase with that rider stripped so the
	// keyword filter does not collapse to the selector's single-keyword slot.
	outputClause := resolving[2 : len(resolving)-4]
	nounClause := outputClause
	for i := range outputClause {
		if equalWord(outputClause[i], "with") {
			nounClause = outputClause[:i]
			break
		}
	}
	tokenPower, tokenToughness, tokenPTKnown := parseTokenPowerToughness(EffectCreate, outputClause)
	identityEffect := EffectSyntax{
		Kind:                EffectCreate,
		Context:             EffectContextReferencedObject,
		Span:                shared.SpanOf(resolving),
		VerbSpan:            resolving[len(resolving)-3].Span,
		ClauseSpan:          shared.SpanOf(resolving),
		Text:                sentence.Text,
		Tokens:              append([]shared.Token(nil), resolving...),
		Selection:           parseSelection(nounClause, atoms),
		Amount:              EffectAmountSyntax{Value: 1, Known: true},
		TokenPower:          tokenPower,
		TokenToughness:      tokenToughness,
		TokenPTKnown:        tokenPTKnown,
		TokenKeywords:       parseTokenKeywords(EffectCreate, outputClause, atoms),
		TokenToxic:          parseTokenKeywordToxic(EffectCreate, outputClause, atoms),
		TokenName:           parseTokenName(EffectCreate, nounClause),
		TokenPredefinedName: parsePredefinedTokenName(EffectCreate, nounClause),
		Replacement: EffectReplacementSyntax{
			Kind: EffectReplacementThatManyIdentity,
			Span: resolving[0].Span,
		},
		References: referencesInSpan(atoms, shared.SpanOf(resolving)),
	}
	if conjunctiveTypeTarget(identityEffect.Selection) {
		identityEffect.Selection.ConjunctiveTypes = true
	}
	return []EffectSyntax{createEffect, identityEffect}, true
}

// matchPassiveTokenIdentity reports the index of the comma separating the
// would-create condition clause from the identity-substitution output when
// tokens spell the passive identity token-creation replacement. The
// controller-only wording ("...under your control, ...") and the
// controller-agnostic wording ("...would be created, ...") are distinguished by
// anyController. The optional card-type word(s) between "more" and "tokens" are
// tolerated here and carried downstream by the would-create group's selector,
// mirroring the additive form.
func matchPassiveTokenIdentity(tokens []shared.Token) (commaIndex int, anyController, ok bool) {
	if !effectWordsAt(tokens, 0, "if", "one", "or", "more") {
		return 0, false, false
	}
	for i := range tokens {
		if tokens[i].Kind != shared.Comma || !effectWordsAt(tokens, i+1, "that", "many") {
			continue
		}
		body := tokens[1:i]
		controllerOnly := false
		if _, stripped := stripTokenSuffix(body, "tokens", "would", "be", "created", "under", "your", "control"); stripped {
			controllerOnly = true
		} else if _, stripped := stripTokenSuffix(body, "tokens", "would", "be", "created"); !stripped {
			continue
		}
		if !effectWordsAt(body, 0, "one", "or", "more") {
			continue
		}
		resolving := tokens[i+1:]
		n := len(resolving)
		// Reject the "that many of those tokens" multiplier wording and the "those
		// tokens plus that many" additive wording, which are not identity
		// substitutions; both are handled by their own recognizers.
		if effectWordsAt(resolving, 2, "of", "those", "tokens") {
			continue
		}
		if n < 7 ||
			!effectWordsAt(resolving, n-4, "are", "created", "instead") ||
			resolving[n-1].Kind != shared.Period {
			continue
		}
		return i, !controllerOnly, true
	}
	return 0, false, false
}
