package compiler

import "github.com/natefinch/council4/mtg/game/zone"

func (reference CompiledReference) OriginalTargetOccurrence() (int, bool) {
	return reference.Subject.targetOccurrence, reference.SubjectSupported() && reference.Binding == ReferenceBindingTarget
}

// DepartureCharacteristicProducer proves the exact move preceding a card-domain
// read. The move captures information, not a new card or permanent subject.
func (reference CompiledReference) DepartureCharacteristicProducer(effects []CompiledEffect, consumer CompiledEffect) (int, bool) {
	if !reference.SupportsUse(ReferenceUseCharacteristic) ||
		reference.SubjectDomain() != ReferenceSubjectCard || reference.Binding != ReferenceBindingTarget ||
		reference.NodeID != consumer.Amount.ReferenceNodeID || reference.Subject.scope != consumer.subjectScope {
		return 0, false
	}
	producer := -1
	consumerFound := false
	for i, effect := range effects {
		if effect.ClauseID == consumer.ClauseID {
			consumerFound = true
			break
		}
		if effect.subjectScope != reference.Subject.scope {
			return 0, false
		}
		for _, target := range effect.Targets {
			if target.Order != reference.Subject.targetOrder {
				continue
			}
			if (effect.Kind != EffectReturn && effect.Kind != EffectPut) ||
				!effect.Exact || effect.Negated || effect.DelayedTiming != 0 ||
				(target.Selector.Zone != zone.Graveyard && target.Selector.Zone != zone.Exile) ||
				(effect.FromZone != zone.None && effect.FromZone != target.Selector.Zone) ||
				effect.ToZone != zone.Hand || len(effect.Targets) != 1 ||
				target.Cardinality.Min != 1 || target.Cardinality.Max != 1 {
				return 0, false
			}
			if producer >= 0 {
				return 0, false
			}
			producer = i
		}
	}
	return producer, consumerFound && producer >= 0
}
