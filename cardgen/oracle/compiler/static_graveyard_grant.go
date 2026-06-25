package compiler

import "github.com/natefinch/council4/cardgen/oracle/parser"

// recognizeStaticGraveyardCardKeywordGrantDeclaration maps the parser's
// graveyard keyword-grant syntax ("[During your turn,] <filter> cards in your
// graveyard have <keyword>.", Six, Wrenn and Six Emblem) onto a closed semantic
// declaration. The granted keyword is the ability's single recognized
// parameterless keyword; the affected set is the controller's matching graveyard
// cards. The ability must be a bare static shell carrying no other semantic
// content.
func recognizeStaticGraveyardCardKeywordGrantDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationGraveyardCardKeywordGrant) {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	if !staticGraveyardKeywordGrantGatingHolds(ability) {
		return StaticDeclaration{}, false
	}
	if !staticGraveyardCardFilterSupported(node.Subject.CardFilter) {
		return StaticDeclaration{}, false
	}
	keyword := ability.Content.Keywords[0]
	return StaticDeclaration{
		Kind:          StaticDeclarationGraveyardCardKeywordGrant,
		Span:          node.Span,
		OperationSpan: keyword.Span,
		Group: StaticGroupReference{
			Span:   node.Subject.Span,
			Domain: StaticGroupControllerGraveyardCards,
		},
		GraveyardGrant: &StaticGraveyardKeywordGrantDeclaration{
			Keyword:              keyword,
			Filter:               node.Subject.CardFilter,
			DuringControllerTurn: node.RestrictDuringControllerTurn,
		},
	}, true
}

// staticGraveyardKeywordGrantGatingHolds reports whether the ability is a bare
// static keyword-grant shell whose only resolving content is the single
// "<subject> have <keyword>" grant-keyword effect and the parameterless keyword
// it confers.
func staticGraveyardKeywordGrantGatingHolds(ability CompiledAbility) bool {
	return ability.Cost == nil &&
		ability.Trigger == nil &&
		len(ability.Content.Modes) == 0 &&
		len(ability.Content.Targets) == 0 &&
		len(ability.Content.Conditions) == 0 &&
		len(ability.Content.Effects) == 1 &&
		ability.Content.Effects[0].Kind == EffectGrantKeyword &&
		ability.Content.Effects[0].Duration == DurationNone &&
		len(ability.Content.Keywords) == 1 &&
		ability.Content.Keywords[0].ParameterKind == parser.KeywordParameterNone
}

// staticGraveyardCardFilterSupported reports whether the parsed card filter is
// one the runtime card selection can represent.
func staticGraveyardCardFilterSupported(filter parser.StaticDeclarationCardFilterKind) bool {
	switch filter {
	case parser.StaticDeclarationCardFilterNonlandPermanent,
		parser.StaticDeclarationCardFilterPermanent,
		parser.StaticDeclarationCardFilterCreature,
		parser.StaticDeclarationCardFilterLand,
		parser.StaticDeclarationCardFilterInstantOrSorcery:
		return true
	default:
		return false
	}
}
