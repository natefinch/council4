package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// appendStaticSpellKeywordGrantDeclaration lowers a spell keyword-grant
// declaration ("[<filter>] spells you cast have <keyword>.", Inspiring Statuary,
// Ironheart, Clever Champion) into a RuleEffectGrantSpellKeyword rule effect on
// the static body. The effect confers the parsed cost-affecting keyword on the
// matching spells the controller casts; the payment planner honors it before
// cost payment, so a granted spell can pay with the keyword's machinery in
// addition to any keyword it carries natively.
func appendStaticSpellKeywordGrantDeclaration(body *game.StaticAbility, declaration compiler.StaticDeclaration) bool {
	if declaration.SpellGrant == nil ||
		declaration.Group.Domain != compiler.StaticGroupControllerSpells {
		return false
	}
	grant := declaration.SpellGrant
	keyword, ok := spellCostGrantRuntimeKeyword(grant.Keyword.Kind)
	if !ok {
		return false
	}
	selection, ok := spellKeywordGrantSelection(grant.Filter)
	if !ok {
		return false
	}
	body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
		Kind:               game.RuleEffectGrantSpellKeyword,
		AffectedController: game.ControllerYou,
		CardSelection:      selection,
		GrantedKeyword:     keyword,
	})
	return true
}

// spellKeywordGrantSelection maps the parsed card filter onto the runtime card
// selection that matches the affected spells. The unfiltered form yields an
// empty selection that matches every spell the controller casts.
func spellKeywordGrantSelection(filter parser.StaticDeclarationCardFilterKind) (game.Selection, bool) {
	switch filter {
	case parser.StaticDeclarationCardFilterNone:
		return game.Selection{}, true
	case parser.StaticDeclarationCardFilterNonartifact:
		return game.Selection{ExcludedTypes: []types.Card{types.Artifact}}, true
	case parser.StaticDeclarationCardFilterNoncreature:
		return game.Selection{ExcludedTypes: []types.Card{types.Creature}}, true
	default:
		return game.Selection{}, false
	}
}

// spellCostGrantRuntimeKeyword maps a parser keyword kind onto the runtime
// cost-affecting keyword a spell keyword grant confers. Only the cost-reducing
// keywords Improvise, Convoke, and Delve are supported; any other keyword fails
// closed so no grant the payment machinery ignores is ever generated.
func spellCostGrantRuntimeKeyword(kind parser.KeywordKind) (game.Keyword, bool) {
	switch kind {
	case parser.KeywordImprovise:
		return game.Improvise, true
	case parser.KeywordConvoke:
		return game.Convoke, true
	case parser.KeywordDelve:
		return game.Delve, true
	default:
		return game.KeywordNone, false
	}
}
