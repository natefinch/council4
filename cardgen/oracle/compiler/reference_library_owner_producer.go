package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/zone"
)

func exactLibraryCardReferenceProducer(effect CompiledEffect, effects []CompiledEffect) bool {
	if SingularLibraryCardProducer(effect) {
		return true
	}
	if effect.LibraryOwnerClauseID <= 0 || effect.Kind != EffectReveal ||
		effect.Context != parser.EffectContextPriorSubject || effect.Player != parser.EffectPlayerTargetOwner ||
		effect.CardSource != parser.EffectCardSourceTopOfPlayerLibrary ||
		!effect.Exact || effect.Negated || effect.Optional || effect.DelayedTiming != 0 ||
		!effect.Amount.Known || effect.Amount.Value != 1 || len(effect.Targets) != 0 {
		return false
	}
	owners := 0
	for i, owner := range effects {
		if owner.ClauseID != effect.LibraryOwnerClauseID {
			continue
		}
		if i+1 >= len(effects) || effects[i+1].ClauseID != effect.ClauseID ||
			owner.Kind != EffectShuffle || !owner.Exact || owner.Negated || owner.Optional ||
			owner.Context != parser.EffectContextTarget || owner.Player != parser.EffectPlayerTargetOwner ||
			owner.ToZone != zone.Library || len(owner.Targets) != 1 ||
			owner.Targets[0].Selector.Kind != SelectorPermanent ||
			owner.Targets[0].Cardinality.Min != 1 || owner.Targets[0].Cardinality.Max != 1 {
			return false
		}
		owners++
	}
	return owners == 1
}
