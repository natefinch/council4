package cardgen

import (
	"fmt"
	"strconv"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// appendStaticGraveyardCardKeywordGrantDeclaration lowers a graveyard
// keyword-grant declaration ("[During your turn,] <filter> cards in your
// graveyard have <keyword>.", Six, Wrenn and Six Emblem; "Each nonland card in
// your graveyard has escape. The escape cost is equal to the card's mana cost
// plus exile N other cards from your graveyard.", Underworld Breach) into a
// RuleEffectGrantGraveyardCardKeyword rule effect on the static body. The effect
// grants the keyword to the controller's matching graveyard cards, carrying the
// escape variant's computed cast cost so the runtime can synthesize each card's
// escape alternative.
func appendStaticGraveyardCardKeywordGrantDeclaration(body *game.StaticAbility, declaration compiler.StaticDeclaration) bool {
	if declaration.Group.Domain != compiler.StaticGroupControllerGraveyardCards {
		return false
	}
	grant := declaration.GraveyardGrant
	keyword, castCost, ok := graveyardGrantKeywordAndCost(grant)
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
		GraveyardCastCost:              castCost,
	})
	return true
}

// graveyardGrantKeywordAndCost resolves the runtime keyword and computed cast
// cost a graveyard keyword-grant declaration confers. The escape variant
// (Underworld Breach) carries a typed EscapeCost the parser recognized, which
// lowers to the escaping card's own mana cost plus an
// exile-N-other-cards-from-your-graveyard additional cost that excludes the
// escaping card itself. Every other grant is a parameterless keyword
// (retrace-style) whose cost the keyword itself defines, so it carries no
// computed cost. It fails closed on an escape grant missing its computed cost and
// on any keyword the runtime does not model.
func graveyardGrantKeywordAndCost(grant *compiler.StaticGraveyardKeywordGrantDeclaration) (game.Keyword, game.GraveyardCastGrantCost, bool) {
	if grant.EscapeCost != nil {
		if grant.Keyword.Kind != parser.KeywordEscape || !grant.EscapeCost.UseCardManaCost || grant.EscapeCost.ExileOtherCount < 1 {
			return game.KeywordNone, game.GraveyardCastGrantCost{}, false
		}
		count := grant.EscapeCost.ExileOtherCount
		return game.Escape, game.GraveyardCastGrantCost{
			UseCardManaCost: true,
			AdditionalCosts: []cost.Additional{{
				Kind:          cost.AdditionalExile,
				Text:          graveyardEscapeExileText(count),
				Source:        zone.Graveyard,
				Amount:        count,
				ExcludeSource: true,
			}},
		}, true
	}
	keyword, ok := runtimeKeyword(grant.Keyword.Kind)
	if !ok {
		return game.KeywordNone, game.GraveyardCastGrantCost{}, false
	}
	return keyword, game.GraveyardCastGrantCost{}, true
}

// graveyardEscapeExileText renders the "Exile N other cards from your graveyard"
// additional-cost text in the same word-cardinal style native escape costs use,
// so a granted escape cost reads like a printed one in logs and generated source.
func graveyardEscapeExileText(count int) string {
	noun := "cards"
	if count == 1 {
		noun = "card"
	}
	return fmt.Sprintf("Exile %s other %s from your graveyard", cardinalWord(count), noun)
}

// cardinalWord returns the English cardinal word for a small non-negative count
// (matching the parser's cardinal atoms), falling back to the decimal form for
// values outside the named range.
func cardinalWord(n int) string {
	words := []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten"}
	if n >= 0 && n < len(words) {
		return words[n]
	}
	return strconv.Itoa(n)
}

// graveyardCardKeywordGrantSelection maps the parsed card filter onto the
// runtime card selection that matches the affected graveyard cards.
func graveyardCardKeywordGrantSelection(filter parser.StaticDeclarationCardFilterKind) (game.Selection, bool) {
	switch filter {
	case parser.StaticDeclarationCardFilterNonland:
		return game.Selection{
			ExcludedTypes: []types.Card{types.Land},
		}, true
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
