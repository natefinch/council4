package compiler

import "github.com/natefinch/council4/cardgen/oracle/parser"

func bindLibraryCardReference(reference *CompiledReference, effects []CompiledEffect) bool {
	if reference.ProducerClauseID <= 0 {
		return false
	}
	if reference.Binding != ReferenceBindingUnsupported &&
		reference.Binding != ReferenceBindingPriorInstructionResult &&
		reference.Binding != ReferenceBindingLibraryOwner {
		return false
	}
	reference.Binding = ReferenceBindingUnsupported
	prior := -1
	for i, effect := range effects {
		if effect.ClauseID != reference.ProducerClauseID {
			continue
		}
		if prior >= 0 || !SingularLibraryCardProducer(effect) {
			return true
		}
		prior = i
	}
	if prior >= 0 {
		reference.Binding = ReferenceBindingPriorInstructionResult
		if reference.Kind == ReferenceThatPlayer {
			if effects[prior].Context != parser.EffectContextTarget {
				reference.Binding = ReferenceBindingUnsupported
				return true
			}
			reference.Binding = ReferenceBindingLibraryOwner
		}
		reference.PriorInstruction = prior
	}
	return true
}

// SingularLibraryCardProducer proves the modeled exact observation domain.
func SingularLibraryCardProducer(effect CompiledEffect) bool {
	if (effect.Kind != EffectLookAtLibraryTop && effect.Kind != EffectReveal) ||
		effect.CardSource != parser.EffectCardSourceTopOfPlayerLibrary ||
		!effect.Exact || effect.Negated || effect.DelayedTiming != 0 ||
		!effect.Amount.Known || effect.Amount.Value != 1 {
		return false
	}
	switch effect.Context {
	case parser.EffectContextController:
		return len(effect.Targets) == 0
	case parser.EffectContextTarget:
		return len(effect.Targets) == 1 &&
			(effect.Targets[0].Selector.Kind == SelectorPlayer || effect.Targets[0].Selector.Kind == SelectorOpponent) &&
			effect.Targets[0].Cardinality.Min == 1 && effect.Targets[0].Cardinality.Max == 1
	default:
		return false
	}
}
