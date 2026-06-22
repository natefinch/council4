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

// staticGraveyardCardFilter matches the card-type filter words preceding "cards
// in your graveyard" and returns the recognized filter together with the number
// of tokens it consumed.
func staticGraveyardCardFilter(tokens []shared.Token, index int) (StaticDeclarationCardFilterKind, int, bool) {
	switch {
	case staticWordsAt(tokens, index, "nonland", "permanent"):
		return StaticDeclarationCardFilterNonlandPermanent, 2, true
	case staticWordsAt(tokens, index, "instant", "and", "sorcery"):
		return StaticDeclarationCardFilterInstantOrSorcery, 3, true
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
