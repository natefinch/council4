package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func searchReferenceNounCompatible(reference CompiledReference, search CompiledEffect) bool {
	if reference.SubjectNoun == parser.ObjectNounUnknown || reference.SubjectNoun == parser.ObjectNounCard ||
		reference.SubjectNoun == parser.ObjectNounPermanent {
		return true
	}
	cardType, ok := costPermanentTypeFromNoun(reference.SubjectNoun)
	if !ok {
		return false
	}
	required := search.Selector.RequiredTypesAny()
	if len(required) != 0 {
		return len(required) == 1 && required[0] == cardType
	}
	intrinsic := map[SelectorKind]types.Card{
		SelectorArtifact: types.Artifact, SelectorCreature: types.Creature,
		SelectorEnchantment: types.Enchantment, SelectorLand: types.Land,
		SelectorPlaneswalker: types.Planeswalker,
	}
	return intrinsic[search.Selector.Kind] == cardType
}

// A compound singular search/placement still owns the search's product. A later
// permanent consumer must prove the placement, not infer entry from the noun.
func searchedPermanentProduct(reference CompiledReference, effects []CompiledEffect) bool {
	if reference.Binding != ReferenceBindingPriorInstructionResult || reference.PriorInstruction < 0 ||
		reference.PriorInstruction >= len(effects) || reference.CardIdentity {
		return false
	}
	search := effects[reference.PriorInstruction]
	end, ok := singularSearchEntryEnd(search, effects)
	return ok && reference.ProducerClauseID == search.ClauseID &&
		reference.Order.Start >= effects[end-1].VerbOrder.End
}

func singularSearchedPermanentProducer(search CompiledEffect, effects []CompiledEffect) bool {
	_, ok := singularSearchEntryEnd(search, effects)
	return ok
}

func singularSearchEntryEnd(search CompiledEffect, effects []CompiledEffect) (int, bool) {
	if search.Kind != EffectSearch || !search.Exact || search.Negated || search.DelayedTiming != 0 ||
		!search.Amount.Known || search.Amount.Value != 1 || search.Amount.AnyNumber ||
		search.Context != parser.EffectContextController || len(search.Targets) != 0 || search.ClauseID <= 0 {
		return 0, false
	}
	start := -1
	for i, effect := range effects {
		if effect.ClauseID == search.ClauseID {
			if start >= 0 {
				return 0, false
			}
			start = i
		}
	}
	if start < 0 {
		return 0, false
	}
	putIndex := start + 1
	if putIndex < len(effects) && effects[putIndex].Kind == EffectReveal {
		putIndex++
	}
	if putIndex+1 >= len(effects) {
		return 0, false
	}
	put, shuffle := effects[putIndex], effects[putIndex+1]
	if put.Kind != EffectPut || put.ToZone != zone.Battlefield ||
		put.Negated || put.DelayedTiming != 0 || len(put.Targets) != 0 || len(put.References) != 1 ||
		shuffle.Kind != EffectShuffle || shuffle.Connection != parser.EffectConnectionThen {
		return 0, false
	}
	return putIndex + 2, true
}
