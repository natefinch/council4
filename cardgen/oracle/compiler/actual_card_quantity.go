package compiler

import (
	"reflect"
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/zone"
)

type actualCardQuantityProof struct {
	outcome            EffectKind
	consumer, producer int
	multiplier, addend int
	form               DynamicAmountForm
	selector           CompiledSelector
	from, to           zone.Type
	cardSource         parser.EffectCardSourceKind
	producerContext    parser.EffectContextKind
	consumerKind       EffectKind
	consumerContext    parser.EffectContextKind
	issued             bool
}

func bindActualCardQuantities(effects []CompiledEffect) {
	for ci := range effects {
		amount := &effects[ci].Amount
		proof := amount.actualCardResult
		if amount.DynamicKind != DynamicAmountActualCardResultCount || proof == nil ||
			amount.ProducerClauseID <= 0 || amount.selector == nil {
			continue
		}
		producer := -1
		for pi := range effects {
			if effects[pi].ClauseID == amount.ProducerClauseID {
				if producer >= 0 {
					producer = -1
					break
				}
				producer = pi
			}
		}
		if producer < 0 || producer >= ci || effects[ci].ClauseID <= 0 {
			continue
		}
		prior := effects[producer]
		if prior.Kind != proof.outcome || !prior.Exact || prior.Negated || prior.DelayedTiming != 0 ||
			prior.Replacement.Kind != parser.EffectReplacementNone ||
			effects[ci].DelayedTiming != 0 || !effects[ci].Exact || effects[ci].Negated {
			continue
		}
		proof.consumer, proof.producer = effects[ci].ClauseID, prior.ClauseID
		proof.multiplier, proof.addend, proof.form = amount.Multiplier, amount.Addend, amount.DynamicForm
		proof.selector = cloneActualCardSelector(*amount.selector)
		proof.from, proof.to = prior.FromZone, prior.ToZone
		proof.cardSource, proof.producerContext = prior.CardSource, prior.Context
		proof.consumerKind, proof.consumerContext = effects[ci].Kind, effects[ci].Context
		proof.issued = true
	}
}

func cloneActualCardSelector(selector CompiledSelector) CompiledSelector {
	if selector.atoms == nil {
		return selector
	}
	atoms := selectorAtoms(selector)
	atoms.RequiredTypesAny = slices.Clone(atoms.RequiredTypesAny)
	atoms.ExcludedTypes = slices.Clone(atoms.ExcludedTypes)
	atoms.Supertypes = slices.Clone(atoms.Supertypes)
	atoms.ExcludedSupertypes = slices.Clone(atoms.ExcludedSupertypes)
	atoms.ColorsAny = slices.Clone(atoms.ColorsAny)
	atoms.ExcludedColors = slices.Clone(atoms.ExcludedColors)
	atoms.SubtypesAny = slices.Clone(atoms.SubtypesAny)
	atoms.ExcludedSubtypes = slices.Clone(atoms.ExcludedSubtypes)
	atoms.SourceTypes = slices.Clone(atoms.SourceTypes)
	return selector.WithAtoms(atoms)
}

// ActualCardQuantitySupported requires the compiler-issued, frozen quantity proof.
func (a CompiledAmount) ActualCardQuantitySupported() bool {
	proof := a.actualCardResult
	return a.DynamicKind == DynamicAmountActualCardResultCount && proof != nil && proof.issued &&
		a.ProducerClauseID == proof.producer && a.Multiplier == proof.multiplier &&
		a.Addend == proof.addend && a.DynamicForm == proof.form && a.selector != nil &&
		!a.Known && !a.VariableX && !a.RangeKnown && !a.AnyNumber &&
		reflect.DeepEqual(*a.selector, proof.selector)
}

// ActualCardQuantityProducer revalidates the owned producer after effect planning.
func (a CompiledAmount) ActualCardQuantityProducer(effects []CompiledEffect, consumer int) (int, bool) {
	if !a.ActualCardQuantitySupported() || consumer != a.actualCardResult.consumer {
		return 0, false
	}
	proof := a.actualCardResult
	producer, consumerIndex := -1, -1
	for i, effect := range effects {
		if effect.ClauseID == consumer {
			if consumerIndex >= 0 || effect.Kind != proof.consumerKind ||
				effect.Context != proof.consumerContext || !effect.Exact || effect.Negated ||
				effect.DelayedTiming != 0 {
				return 0, false
			}
			consumerIndex = i
		}
		if effect.ClauseID != proof.producer {
			continue
		}
		if producer >= 0 || effect.Kind != proof.outcome || !effect.Exact || effect.Negated ||
			effect.DelayedTiming != 0 || effect.FromZone != proof.from || effect.ToZone != proof.to ||
			effect.CardSource != proof.cardSource || effect.Context != proof.producerContext ||
			effect.Replacement.Kind != parser.EffectReplacementNone {
			return 0, false
		}
		producer = i
	}
	return producer, producer >= 0 && consumerIndex > producer
}
