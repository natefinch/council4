package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// parseStaticGraveyardCardKeywordGrantDeclaration recognizes the graveyard
// keyword-grant family "[During your turn,] <filter> cards in your graveyard
// have <keyword>." (Six's "During your turn, nonland permanent cards in your
// graveyard have retrace."; Wrenn and Six Emblem's "Instant and sorcery cards in
// your graveyard have retrace."). The declaration grants a parameterless keyword
// to the controller's matching graveyard cards. An optional leading "During your
// turn," scopes the grant to the controller's turn (RestrictDuringControllerTurn).
// The granted keyword atom is carried in KeywordSpans and resolved downstream
// from the ability's recognized keyword content.
func parseStaticGraveyardCardKeywordGrantDeclaration(tokens []shared.Token, atoms Atoms) (StaticDeclarationSyntax, bool) {
	if len(tokens) < 6 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	end := len(tokens) - 1
	index := 0
	duringControllerTurn := false
	if staticWordsAt(tokens, index, "during", "your", "turn") {
		duringControllerTurn = true
		index += 3
		if index < end && tokens[index].Kind == shared.Comma {
			index++
		}
	}
	filter, width, ok := staticGraveyardCardFilter(tokens, index)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	index += width
	if !staticWordsAt(tokens, index, "cards", "in", "your", "graveyard", "have") {
		return StaticDeclarationSyntax{}, false
	}
	index += 5
	keyword, keywordWidth, ok := staticKeywordAt(tokens, index, end, atoms)
	if !ok || keyword.Parameter.Kind != KeywordParameterNone || index+keywordWidth != end {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationGraveyardCardKeywordGrant,
		Span:          shared.SpanOf(tokens),
		OperationSpan: keyword.Span,
		Subject: StaticDeclarationSubject{
			Kind:       StaticDeclarationSubjectControllerGraveyard,
			Span:       shared.SpanOf(tokens[:index]),
			CardFilter: filter,
		},
		KeywordSpans:                 []shared.Span{keyword.Span},
		RestrictDuringControllerTurn: duringControllerTurn,
	}, true
}

// parseStaticGraveyardEscapeGrantDeclaration recognizes the computed-escape
// graveyard grant "[During your turn,] [Each] <filter> card[s] in your graveyard
// has escape. The escape cost is equal to the card's mana cost plus exile N other
// cards from your graveyard." (Underworld Breach). Unlike the parameterless
// graveyard keyword grant, escape carries a per-card computed cost defined by the
// second sentence, captured as GraveyardEscapeCost. The parser accepts only this
// exact shape — the card's own mana cost plus an "exile N other cards from your
// graveyard" additional cost — so any other computed escape wording yields no
// declaration and the card fails closed rather than lowering an approximate cost.
func parseStaticGraveyardEscapeGrantDeclaration(tokens []shared.Token, atoms Atoms) (StaticDeclarationSyntax, bool) {
	if len(tokens) == 0 || tokens[len(tokens)-1].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	firstPeriod := -1
	for i := range tokens {
		if tokens[i].Kind == shared.Period {
			firstPeriod = i
			break
		}
	}
	// The grant and its escape-cost definition are two sentences: the first
	// period must fall strictly inside the body, leaving a non-empty cost sentence.
	if firstPeriod <= 0 || firstPeriod >= len(tokens)-1 {
		return StaticDeclarationSyntax{}, false
	}
	grantTokens := tokens[:firstPeriod+1]
	costTokens := tokens[firstPeriod+1:]

	grant, ok := parseGraveyardEscapeGrantSentence(grantTokens, atoms)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	exileCount, ok := parseGraveyardEscapeCostSentence(costTokens)
	if !ok {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationGraveyardCardKeywordGrant,
		Span:          shared.SpanOf(tokens),
		OperationSpan: grant.Keyword.Span,
		Subject: StaticDeclarationSubject{
			Kind:       StaticDeclarationSubjectControllerGraveyard,
			Span:       shared.SpanOf(grantTokens[:len(grantTokens)-1]),
			CardFilter: grant.Filter,
		},
		KeywordSpans:                 []shared.Span{grant.Keyword.Span},
		RestrictDuringControllerTurn: grant.DuringControllerTurn,
		GraveyardEscapeCost: &StaticGraveyardEscapeCostSyntax{
			UseCardManaCost: true,
			ExileOtherCount: exileCount,
		},
	}, true
}

// graveyardEscapeGrant holds the fields recognized from the first sentence of a
// computed escape grant: the matched card filter, the Escape keyword atom, and
// whether the grant is limited to the controller's turn.
type graveyardEscapeGrant struct {
	Filter               StaticDeclarationCardFilterKind
	Keyword              Keyword
	DuringControllerTurn bool
}

// parseGraveyardEscapeGrantSentence matches the first sentence of a computed
// escape grant, "[During your turn,] [Each] <filter> card[s] in your graveyard
// has escape.", and returns the recognized card filter and Escape keyword atom.
// It requires the granted keyword to be a parameterless escape spanning to the
// sentence's end, so only the escape family reaches the computed-cost parser.
func parseGraveyardEscapeGrantSentence(tokens []shared.Token, atoms Atoms) (graveyardEscapeGrant, bool) {
	end := len(tokens) - 1 // exclude the terminating period
	index := 0
	duringControllerTurn := false
	if staticWordsAt(tokens, index, "during", "your", "turn") {
		duringControllerTurn = true
		index += 3
		if index < end && tokens[index].Kind == shared.Comma {
			index++
		}
	}
	if staticWordsAt(tokens, index, "each") {
		index++
	}
	filter, width, ok := staticGraveyardCardFilter(tokens, index)
	if !ok {
		return graveyardEscapeGrant{}, false
	}
	index += width
	if !staticWordsAt(tokens, index, "card") && !staticWordsAt(tokens, index, "cards") {
		return graveyardEscapeGrant{}, false
	}
	index++
	if !staticWordsAt(tokens, index, "in", "your", "graveyard") {
		return graveyardEscapeGrant{}, false
	}
	index += 3
	if !staticWordsAt(tokens, index, "has") && !staticWordsAt(tokens, index, "have") {
		return graveyardEscapeGrant{}, false
	}
	index++
	keyword, keywordWidth, ok := staticKeywordAt(tokens, index, end, atoms)
	if !ok || keyword.Kind != KeywordEscape || keyword.Parameter.Kind != KeywordParameterNone || index+keywordWidth != end {
		return graveyardEscapeGrant{}, false
	}
	return graveyardEscapeGrant{Filter: filter, Keyword: keyword, DuringControllerTurn: duringControllerTurn}, true
}

// parseGraveyardEscapeCostSentence matches the escape-cost definition "The escape
// cost is equal to the card's mana cost plus exile N other cards from your
// graveyard." and returns N. It accepts only the card's own mana cost plus a
// graveyard-exile additional cost, so any other computed form is rejected.
func parseGraveyardEscapeCostSentence(tokens []shared.Token) (int, bool) {
	if len(tokens) == 0 || tokens[len(tokens)-1].Kind != shared.Period {
		return 0, false
	}
	body := tokens[:len(tokens)-1]
	prefix := []string{"the", "escape", "cost", "is", "equal", "to", "the", "card's", "mana", "cost", "plus", "exile"}
	if !staticWordsAt(body, 0, prefix...) {
		return 0, false
	}
	index := len(prefix)
	if index >= len(body) {
		return 0, false
	}
	count, ok := addendCardinal(body[index])
	if !ok || count < 1 {
		return 0, false
	}
	index++
	if !staticWordsAt(body, index, "other", "cards", "from", "your", "graveyard") {
		return 0, false
	}
	index += 5
	if index != len(body) {
		return 0, false
	}
	return count, true
}

// staticGraveyardCardFilter matches the card-type filter words preceding "cards
// in your graveyard" and returns the recognized filter together with the number
// of tokens it consumed.
func staticGraveyardCardFilter(tokens []shared.Token, index int) (StaticDeclarationCardFilterKind, int, bool) {
	switch {
	case staticWordsAt(tokens, index, "nonland", "permanent"):
		return StaticDeclarationCardFilterNonlandPermanent, 2, true
	case staticWordsAt(tokens, index, "instant", "and", "sorcery"):
		return StaticDeclarationCardFilterInstantOrSorcery, 3, true
	case staticWordsAt(tokens, index, "nonland"):
		return StaticDeclarationCardFilterNonland, 1, true
	case staticWordsAt(tokens, index, "permanent"):
		return StaticDeclarationCardFilterPermanent, 1, true
	case staticWordsAt(tokens, index, "creature"):
		return StaticDeclarationCardFilterCreature, 1, true
	case staticWordsAt(tokens, index, "land"):
		return StaticDeclarationCardFilterLand, 1, true
	default:
		return StaticDeclarationCardFilterNone, 0, false
	}
}
