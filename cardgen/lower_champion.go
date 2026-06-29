package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// lowerChampionAbility lowers the Champion keyword (CR 702.71): "Champion a/an
// <type> (When this enters, sacrifice it unless you exile another <type> you
// control. When this leaves the battlefield, that card returns to the
// battlefield.)" Its enters-the-battlefield half exiles a chosen matching
// permanent under the shared exile-until-leaves key, and the face-level
// synthesizeExileUntilLeavesReturns pass adds the paired return on leave, so
// only the enter trigger is emitted here. The keyword's type filter always
// resolves to creatures the controller owns; "Champion a creature" matches any
// creature and "Champion a Goblin" narrows to that subtype. Only the exact bare
// keyword with its parenthesized reminder is supported.
func lowerChampionAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.TriggeredAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 || ability.Content.Keywords[0].Kind != parser.KeywordChampion {
		return game.TriggeredAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	selection, ok := championSelection(keyword.EnchantTarget)
	if !ok ||
		keyword.ParameterKind != parser.KeywordParameterChampion ||
		(ability.Kind != compiler.AbilityStatic && ability.Kind != compiler.AbilitySpell) ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" ||
		!keywordOnlyCovered(syntax, keyword) {
		return game.TriggeredAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Champion ability",
			"the executable source backend supports only the exact Champion keyword with a creature type",
		)
	}
	return game.TriggeredAbility{
		Text: keyword.Text,
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhen,
			Pattern: game.TriggerPattern{
				Event:  game.EventPermanentEnteredBattlefield,
				Source: game.TriggerSourceSelf,
			},
		},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.ChampionExile{
				Selection: selection,
				LinkedKey: exileUntilLeavesKey,
			},
		}}}.Ability(),
	}, true, nil
}

// championSelection builds the controller-owned creature filter for the
// Champion keyword's type. A bare "creature" matches any creature; a subtype
// list narrows to those creature subtypes. The exiled permanent must be another
// permanent the controller owns, so the selection is anchored to the controller
// and excludes the source. Any non-creature card type fails closed.
func championSelection(target compiler.CompiledEnchantTarget) (game.Selection, bool) {
	if !target.Known || len(target.CardTypes) > 1 {
		return game.Selection{}, false
	}
	for _, cardType := range target.CardTypes {
		if cardType != types.Creature {
			return game.Selection{}, false
		}
	}
	return game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		SubtypesAny:   target.Subtypes,
		Controller:    game.ControllerYou,
		ExcludeSource: true,
	}, true
}
