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

// parseStaticCreatureAttackTaxDeclaration recognizes the per-creature attack-tax
// family that taxes each attacker a per-creature cost. It covers the
// planeswalker-inclusive wording "Creatures can't attack you or planeswalkers
// you control unless their controller pays {COST} for each of those
// creatures[, where X is the number of enchantments you control]." (Baird,
// Archon of Absolution with a fixed {N}; Sphere of Safety with {X} scaled by
// enchantments) and the player-only domain wording "Creatures can't attack you
// unless their controller pays {X} for each creature they control that's
// attacking you, where X is the number of basic land types among lands you
// control." (Collective Restraint). Any deviation from the exact wording leaves
// the clause unconsumed and fails closed.
func parseStaticCreatureAttackTaxDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if declaration, ok := parseStaticPlaneswalkerAttackTaxDeclaration(tokens); ok {
		return declaration, true
	}
	return parseStaticDomainAttackTaxDeclaration(tokens)
}

// parseStaticPlaneswalkerAttackTaxDeclaration recognizes the planeswalker-
// inclusive per-creature attack tax "Creatures can't attack you or planeswalkers
// you control unless their controller pays {COST} for each of those
// creatures[, where X is the number of enchantments you control]." A fixed {N}
// cost ends the sentence (Baird, Archon of Absolution); a {X} cost requires the
// trailing "where X is the number of enchantments you control" clause (Sphere of
// Safety).
func parseStaticPlaneswalkerAttackTaxDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) < 19 ||
		tokens[12].Kind != shared.Symbol ||
		!staticWordsAt(tokens, 0,
			"creatures", "can't", "attack", "you", "or", "planeswalkers", "you", "control",
			"unless", "their", "controller", "pays") ||
		!staticWordsAt(tokens, 13, "for", "each", "of", "those", "creatures") {
		return StaticDeclarationSyntax{}, false
	}
	if amount, ok := staticGenericSymbolValue(tokens[12].Text); ok && amount > 0 {
		if len(tokens) != 19 || tokens[18].Kind != shared.Period {
			return StaticDeclarationSyntax{}, false
		}
		return StaticDeclarationSyntax{
			Kind:                           StaticDeclarationCreatureAttackTax,
			Span:                           shared.SpanOf(tokens),
			OperationSpan:                  shared.SpanOf(tokens),
			AttackTaxAmountKind:            StaticAttackTaxAmountFixed,
			AttackTaxGeneric:               amount,
			AttackTaxIncludesPlaneswalkers: true,
		}, true
	}
	if tokens[12].Text != "{X}" ||
		len(tokens) != 29 ||
		tokens[18].Kind != shared.Comma ||
		tokens[28].Kind != shared.Period ||
		!staticWordsAt(tokens, 19,
			"where", "x", "is", "the", "number", "of", "enchantments", "you", "control") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:                           StaticDeclarationCreatureAttackTax,
		Span:                           shared.SpanOf(tokens),
		OperationSpan:                  shared.SpanOf(tokens),
		AttackTaxAmountKind:            StaticAttackTaxAmountEnchantments,
		AttackTaxIncludesPlaneswalkers: true,
	}, true
}

// parseStaticDomainAttackTaxDeclaration recognizes the player-only domain attack
// tax "Creatures can't attack you unless their controller pays {X} for each
// creature they control that's attacking you, where X is the number of basic
// land types among lands you control." (Collective Restraint). The protected
// defending player is the controller; planeswalkers are not covered.
func parseStaticDomainAttackTaxDeclaration(tokens []shared.Token) (StaticDeclarationSyntax, bool) {
	if len(tokens) != 32 ||
		tokens[8].Text != "{X}" ||
		tokens[17].Kind != shared.Comma ||
		tokens[31].Kind != shared.Period ||
		!staticWordsAt(tokens, 0, "creatures", "can't", "attack", "you", "unless", "their", "controller", "pays") ||
		!staticWordsAt(tokens, 9, "for", "each", "creature", "they", "control", "that's", "attacking", "you") ||
		!staticWordsAt(tokens, 18,
			"where", "x", "is", "the", "number", "of", "basic", "land", "types", "among", "lands", "you", "control") {
		return StaticDeclarationSyntax{}, false
	}
	return StaticDeclarationSyntax{
		Kind:                StaticDeclarationCreatureAttackTax,
		Span:                shared.SpanOf(tokens),
		OperationSpan:       shared.SpanOf(tokens),
		AttackTaxAmountKind: StaticAttackTaxAmountDomain,
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
