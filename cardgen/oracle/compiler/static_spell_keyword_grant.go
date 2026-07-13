package compiler

import "github.com/natefinch/council4/cardgen/oracle/parser"

// recognizeStaticSpellKeywordGrantDeclaration maps the parser's spell
// keyword-grant syntax ("[<filter>] spells you cast have <keyword>.", Inspiring
// Statuary, Ironheart, Clever Champion) onto a closed semantic declaration. The
// granted keyword is the ability's single recognized parameterless keyword; the
// affected set is the spells the controller casts, optionally narrowed by card
// type. The ability must be a bare static shell carrying no other semantic
// content, and the keyword must be one the runtime honors during cost payment.
func recognizeStaticSpellKeywordGrantDeclaration(ability CompiledAbility, statics []parser.StaticDeclarationSyntax) (StaticDeclaration, bool) {
	if !staticSyntaxKindsAre(statics, parser.StaticDeclarationSpellKeywordGrant) {
		return StaticDeclaration{}, false
	}
	node := statics[0]
	if !staticSpellKeywordGrantGatingHolds(ability) {
		return StaticDeclaration{}, false
	}
	if !staticSpellKeywordGrantFilterSupported(node.Subject.CardFilter) {
		return StaticDeclaration{}, false
	}
	keyword := ability.Content.Keywords[0]
	if !spellCostGrantKeywordSupported(keyword.Kind) {
		return StaticDeclaration{}, false
	}
	return StaticDeclaration{
		Kind:          StaticDeclarationSpellKeywordGrant,
		Span:          node.Span,
		OperationSpan: keyword.Span,
		Group: StaticGroupReference{
			Span:   node.Subject.Span,
			Domain: StaticGroupControllerSpells,
		},
		SpellGrant: &StaticSpellKeywordGrantDeclaration{
			Keyword: keyword,
			Filter:  node.Subject.CardFilter,
		},
	}, true
}

// staticSpellKeywordGrantGatingHolds reports whether the ability is a bare static
// keyword-grant shell whose only resolving content is the single "<spells> have
// <keyword>" grant-keyword effect and the parameterless keyword it confers. The
// "spells you cast" subject also surfaces as a benign EffectCast clause (as it
// does for the spell cost modifier and cast-as-though-flash declarations), so
// those are tolerated; any other effect kind fails closed.
func staticSpellKeywordGrantGatingHolds(ability CompiledAbility) bool {
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Keywords) != 1 ||
		ability.Content.Keywords[0].ParameterKind != parser.KeywordParameterNone {
		return false
	}
	grants := 0
	for i := range ability.Content.Effects {
		effect := &ability.Content.Effects[i]
		switch effect.Kind {
		case EffectGrantKeyword:
			if effect.Duration != DurationNone {
				return false
			}
			grants++
		case EffectCast:
			// The "spells you cast" subject verb; represented by the typed static
			// declaration, so it carries no additional semantics here.
		default:
			return false
		}
	}
	return grants == 1
}

// staticSpellKeywordGrantFilterSupported reports whether the parsed card filter
// is one the runtime card selection can represent for the affected spells.
func staticSpellKeywordGrantFilterSupported(filter parser.StaticDeclarationCardFilterKind) bool {
	switch filter {
	case parser.StaticDeclarationCardFilterNone,
		parser.StaticDeclarationCardFilterNonartifact,
		parser.StaticDeclarationCardFilterNoncreature:
		return true
	default:
		return false
	}
}

// spellCostGrantKeywordSupported reports whether keyword is a cost-affecting
// keyword the payment machinery honors when a grant confers it on a spell. Only
// the cost-reducing keywords Improvise, Convoke, and Delve are supported; any
// other keyword fails closed so the ability is reported unsupported rather than
// generating a grant the runtime ignores.
func spellCostGrantKeywordSupported(kind parser.KeywordKind) bool {
	switch kind {
	case parser.KeywordImprovise, parser.KeywordConvoke, parser.KeywordDelve:
		return true
	default:
		return false
	}
}
