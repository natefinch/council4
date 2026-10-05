package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/types"
)

func conditionSubtypeListContinues(tokens []shared.Token, comma int) bool {
	return comma > 0 && comma+1 < len(tokens) &&
		conditionSubtypeListTailHasConnector(tokens, comma+1)
}

func conditionSubtypeListTailHasConnector(tokens []shared.Token, start int) bool {
	for i := start; i < len(tokens); {
		if tokens[i].Kind == shared.Comma {
			i++
			continue
		}
		final := equalWord(tokens[i], "or") || equalWord(tokens[i], "and")
		if final {
			i++
		}
		if i < len(tokens) && (equalWord(tokens[i], "a") || equalWord(tokens[i], "an")) {
			i++
		}
		if i+1 >= len(tokens) || tokens[i].Kind != shared.Word || tokens[i+1].Kind != shared.Comma {
			return false
		}
		if final {
			return true
		}
		i += 2
	}
	return false
}

func parseConditionSubtypeList(tokens []shared.Token, atoms Atoms, selection ConditionSelection) (ConditionSelection, bool) {
	start := 0
	sawOr := false
	allLand := true
	var subtypes []types.Sub
	for i := 0; i <= len(tokens); i++ {
		if i < len(tokens) && tokens[i].Kind != shared.Comma && !equalWord(tokens[i], "or") {
			continue
		}
		if i > start {
			member := tokens[start:i]
			if equalWord(member[0], "a") || equalWord(member[0], "an") {
				member = member[1:]
			}
			subtype, ok := atoms.SubtypeAt(shared.SpanOf(member))
			if !ok {
				return ConditionSelection{}, false
			}
			subtypes = append(subtypes, subtype)
			allLand = allLand && SubtypeMatchesCardType(subtype, CardTypeLand)
		} else if i == len(tokens) || tokens[i].Kind == shared.Comma {
			return ConditionSelection{}, false
		}
		if i < len(tokens) && equalWord(tokens[i], "or") {
			if sawOr {
				return ConditionSelection{}, false
			}
			sawOr = true
		} else if sawOr && i < len(tokens) {
			return ConditionSelection{}, false
		}
		start = i + 1
	}
	if !sawOr || len(subtypes) < 3 {
		return ConditionSelection{}, false
	}
	if allLand {
		selection.RequiredTypes = append(selection.RequiredTypes, TriggerCardTypeLand)
	}
	selection.SubtypesAny = append(selection.SubtypesAny, subtypes...)
	return selection, true
}
