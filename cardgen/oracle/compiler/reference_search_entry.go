package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/zone"
)

// A compound singular search/placement still owns the search's product. A later
// permanent consumer must prove the placement, not infer entry from the noun.
func searchedPermanentProduct(reference CompiledReference, effects []CompiledEffect) bool {
	if reference.Binding != ReferenceBindingPriorInstructionResult || reference.PriorInstruction < 0 ||
		reference.PriorInstruction >= len(effects) || reference.CardIdentity {
		return false
	}
	search := effects[reference.PriorInstruction]
	if search.Kind != EffectSearch || !search.Exact || search.Negated ||
		!search.Amount.Known || search.Amount.Value != 1 || len(search.Targets) != 0 {
		return false
	}
	entered := false
	for _, effect := range effects[reference.PriorInstruction+1:] {
		if effectOwnsReference(effect, reference) {
			return entered && (effect.Kind == EffectUntap || effect.Kind == EffectModifyPT ||
				effect.Kind == EffectPut || effect.Kind == EffectGain)
		}
		if effect.Kind == EffectShuffle {
			continue
		}
		if effect.Kind != EffectPut || effect.ToZone != zone.Battlefield || effect.Negated ||
			effect.DelayedTiming != 0 || len(effect.Targets) != 0 || len(effect.References) != 1 ||
			effect.References[0].ProducerClauseID != search.ClauseID ||
			effect.References[0].Binding != ReferenceBindingPriorInstructionResult ||
			effect.References[0].PriorInstruction != reference.PriorInstruction ||
			(effect.References[0].Kind != ReferencePronoun || effect.References[0].Pronoun != ReferencePronounIt) ||
			effect.Context != parser.EffectContextController && effect.Context != parser.EffectContextPriorSubject {
			return false
		}
		if entered {
			return false
		}
		entered = true
	}
	return false
}
