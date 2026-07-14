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

// recognizeStaticGraveyardEscapeGrantDeclaration maps the parser's computed
// escape grant ("Each nonland card in your graveyard has escape. The escape cost
// is equal to the card's mana cost plus exile N other cards from your
// graveyard.", Underworld Breach) onto a graveyard keyword-grant declaration
// carrying the typed EscapeCost. The escape-cost sentence contributes spurious
// legacy grant-keyword and exile effects plus a second escape keyword atom that
// the typed declaration already subsumes, so this recognizer trusts the
// declaration's GraveyardEscapeCost and gates only on the ability being a bare
// escape-grant shell.
func recognizeStaticGraveyardEscapeGrantDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationGraveyardCardKeywordGrant) {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	escape := node.GraveyardEscapeCost
	if escape == nil || !escape.UseCardManaCost || escape.ExileOtherCount < 1 {
		return StaticDeclaration{}, false
	}
	if !staticGraveyardCardFilterSupported(node.Subject.CardFilter) {
		return StaticDeclaration{}, false
	}
	if !staticGraveyardEscapeGrantGatingHolds(ability) {
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
			EscapeCost: &StaticGraveyardEscapeCost{
				UseCardManaCost: escape.UseCardManaCost,
				ExileOtherCount: escape.ExileOtherCount,
			},
		},
	}, true
}

// staticGraveyardEscapeGrantGatingHolds reports whether the ability is the bare
// static shell of a computed escape grant. The escape-cost sentence contributes
// legacy grant-keyword and exile effects and a second escape keyword atom that
// the typed declaration already models, so the gate permits those artifacts
// while still rejecting any ability that carries a cost, trigger, modes, targets,
// conditions, a non-escape keyword, or an unrelated effect.
func staticGraveyardEscapeGrantGatingHolds(ability CompiledAbility) bool {
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) == 0 {
		return false
	}
	for i := range ability.Content.Keywords {
		keyword := ability.Content.Keywords[i]
		if keyword.Kind != parser.KeywordEscape || keyword.ParameterKind != parser.KeywordParameterNone {
			return false
		}
	}
	sawGrant := false
	for i := range ability.Content.Effects {
		switch ability.Content.Effects[i].Kind {
		case EffectGrantKeyword:
			sawGrant = true
		case EffectExile:
			// The "exile N other cards" cost clause yields a legacy exile effect
			// the typed declaration already models; the declaration is authoritative.
		default:
			return false
		}
	}
	return sawGrant
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
		parser.StaticDeclarationCardFilterNonland,
		parser.StaticDeclarationCardFilterPermanent,
		parser.StaticDeclarationCardFilterCreature,
		parser.StaticDeclarationCardFilterLand,
		parser.StaticDeclarationCardFilterInstantOrSorcery:
		return true
	default:
		return false
	}
}
