package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerReconfigureAbility lowers a printed Reconfigure keyword (CR 702.151) to
// its canonical sorcery-speed attach activated ability. It mirrors
// lowerEquipAbility: only an exact "Reconfigure {cost}" keyword with a mana cost
// and no other rules text is supported. The em-dash and ability-word forms, the
// unattach mode, and the "while attached, this isn't a creature" type-change are
// not yet lowered, so it fails closed for any deviation.
func lowerReconfigureAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.ActivatedAbility, bool, *shared.Diagnostic) {
	if len(ability.Content.Keywords) != 1 ||
		ability.Content.Keywords[0].Kind != parser.KeywordReconfigure ||
		ability.Kind != compiler.AbilityStatic {
		return game.ActivatedAbility{}, false, nil
	}
	keyword := ability.Content.Keywords[0]
	if keyword.ParameterKind != parser.KeywordParameterManaCost ||
		ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Targets) != 0 ||
		len(ability.Content.Conditions) != 0 ||
		len(ability.Content.Effects) != 0 ||
		len(ability.Content.References) != 0 ||
		ability.AbilityWord != "" {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Reconfigure ability",
			"the executable source backend supports only exact Reconfigure with a mana cost",
		)
	}
	if len(keyword.ManaCost) == 0 {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Reconfigure ability",
			"the executable source backend supports only exact Reconfigure with a mana cost",
		)
	}
	if keyword.EquipRestriction != nil {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Reconfigure ability",
			"the executable source backend supports only exact Reconfigure with a mana cost",
		)
	}
	if !keywordOnlyCovered(syntax, keyword) {
		return game.ActivatedAbility{}, true, executableDiagnostic(
			ability,
			"unsupported Reconfigure ability",
			"the executable source backend supports only exact Reconfigure with a mana cost",
		)
	}
	return game.ReconfigureActivatedAbility(slices.Clone(keyword.ManaCost)), true, nil
}
