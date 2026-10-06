package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/zone"
)

func bindOwnedProducerReference(reference *CompiledReference, effects []CompiledEffect) bool {
	if reference.LibraryCardObservation || reference.Kind == ReferenceThatPlayer {
		return bindLibraryCardReference(reference, effects)
	}
	if reference.ProducerClauseID <= 0 {
		return false
	}
	reference.Binding = ReferenceBindingUnsupported
	index := -1
	for i, effect := range effects {
		if effect.ClauseID != reference.ProducerClauseID {
			continue
		}
		if index >= 0 || !singularEnteredSubjectProducer(effect, effects) {
			return true
		}
		index = i
	}
	if index >= 0 {
		reference.Binding = ReferenceBindingPriorInstructionResult
		reference.PriorInstruction = index
	}
	return true
}

func singularEnteredSubjectProducer(effect CompiledEffect, effects []CompiledEffect) bool {
	if effect.Negated || effect.DelayedTiming != 0 ||
		(effect.Kind != EffectPut && effect.Kind != EffectReturn) || effect.ToZone != zone.Battlefield {
		return false
	}
	if len(effect.Targets) == 0 {
		for i := range effects {
			if effects[i].ClauseID == effect.ClauseID && singularReturnedPermanentProducer(effects, i) {
				return true
			}
		}
	}
	if !effect.Exact {
		return false
	}
	if len(effect.Targets) == 1 {
		target := effect.Targets[0]
		return targetSubjectDomain(target) == ReferenceSubjectCard && target.Selector.Zone != zone.None &&
			target.Selector.Zone != zone.Battlefield &&
			target.Cardinality.Min == 1 && target.Cardinality.Max == 1
	}
	if len(effect.Targets) != 0 {
		return false
	}
	if effect.CardSource == parser.EffectCardSourcePriorInstructionResult && effect.FromZone == zone.Library {
		sources := 0
		for _, reference := range effect.References {
			if reference.ProducerClauseID <= 0 || !reference.LibraryCardObservation {
				continue
			}
			found := 0
			for _, producer := range effects {
				if producer.ClauseID == reference.ProducerClauseID && exactLibraryCardReferenceProducer(producer, effects) {
					found++
				}
			}
			if found != 1 {
				return false
			}
			sources++
		}
		return sources == 1
	}
	return false
}
