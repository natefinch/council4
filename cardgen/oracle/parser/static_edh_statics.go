package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// parseStaticOpeningHandPlayDeclaration recognizes the pre-game permission "If
// this card is in your opening hand, you may begin the game with it on the
// battlefield." (the Leyline cycle). The permission is a special action taken
// before the game begins; this engine starts every game from a fixed setup and
// never models opening hands, so the declaration carries no runtime payload and
// lowers to an inert static ability. Any deviation from the exact wording leaves
// the clause unconsumed and fails closed.
func parseStaticOpeningHandPlayDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 20 ||
		tokens[8].Kind != shared.Comma ||
		tokens[19].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0, "if", "this", "card", "is", "in", "your", "opening", "hand") ||
		!staticWordsAt(tokens, 9, "you", "may", "begin", "the", "game", "with", "it", "on", "the", "battlefield") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationOpeningHandPlay,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens),
	}, true
}

// parseStaticOpponentEnteringTriggerSuppressionDeclaration recognizes the static
// "Permanents entering don't cause abilities of permanents your opponents
// control to trigger." (Elesh Norn, Mother of Machines). It suppresses the
// entering-caused triggered abilities of permanents the controller's opponents
// control. The declaration carries fixed semantics; any deviation leaves the
// clause unconsumed and fails closed.
func parseStaticOpponentEnteringTriggerSuppressionDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 13 || tokens[12].Kind != shared.Period {
		return StaticDeclarationSyntax{}, false
	}
	if !staticWordsAt(tokens, 0,
		"permanents", "entering", "don't", "cause", "abilities", "of",
		"permanents", "your", "opponents", "control", "to", "trigger") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:          StaticDeclarationOpponentEnteringTriggerSuppression,
		Span:          shared.SpanOf(tokens),
		OperationSpan: shared.SpanOf(tokens),
	}, true
}
