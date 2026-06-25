package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// appendStaticGraveyardCardKeywordGrantDeclaration lowers a graveyard
// keyword-grant declaration ("[During your turn,] <filter> cards in your
// graveyard have <keyword>.", Six, Wrenn and Six Emblem) into a
// RuleEffectGrantGraveyardCardKeyword rule effect on the static body. The effect
// grants the parsed keyword to the controller's matching graveyard cards.
func appendStaticGraveyardCardKeywordGrantDeclaration(body *game.StaticAbility, declaration compiler.StaticDeclaration) bool {
	if declaration.Group.Domain != compiler.StaticGroupControllerGraveyardCards {
		return false
	}
	grant := declaration.GraveyardGrant
	keyword, ok := runtimeKeyword(grant.Keyword.Kind)
	if !ok {
		return false
	}
	selection, ok := graveyardCardKeywordGrantSelection(grant.Filter)
	if !ok {
		return false
	}
	body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
		Kind:                           game.RuleEffectGrantGraveyardCardKeyword,
		AffectedPlayer:                 game.PlayerYou,
		CardSelection:                  selection,
		GrantedKeyword:                 keyword,
		RestrictedDuringControllerTurn: grant.DuringControllerTurn,
	})
	return true
}

// graveyardCardKeywordGrantSelection maps the parsed card filter onto the
// runtime card selection that matches the affected graveyard cards.
func graveyardCardKeywordGrantSelection(filter parser.StaticDeclarationCardFilterKind) (game.Selection, bool) {
	switch filter {
	case parser.StaticDeclarationCardFilterNonlandPermanent:
		return game.Selection{
			RequiredTypesAny: []types.Card{types.Creature, types.Artifact, types.Enchantment, types.Planeswalker, types.Battle},
			ExcludedTypes:    []types.Card{types.Land},
		}, true
	case parser.StaticDeclarationCardFilterPermanent:
		return game.Selection{
			RequiredTypesAny: []types.Card{types.Land, types.Creature, types.Artifact, types.Enchantment, types.Planeswalker, types.Battle},
		}, true
	case parser.StaticDeclarationCardFilterCreature:
		return game.Selection{RequiredTypes: []types.Card{types.Creature}}, true
	case parser.StaticDeclarationCardFilterLand:
		return game.Selection{RequiredTypes: []types.Card{types.Land}}, true
	case parser.StaticDeclarationCardFilterInstantOrSorcery:
		return game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}}, true
	default:
		return game.Selection{}, false
	}
}
