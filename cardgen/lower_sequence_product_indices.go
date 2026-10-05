package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
)

// Clipping and folding can change effect indices, but never producer ClauseIDs.
func normalizeSequenceProductReferences(content compiler.AbilityContent) (compiler.AbilityContent, bool) {
	indices := make(map[int]int)
	for index, effect := range content.Effects {
		if effect.ClauseID <= 0 {
			continue
		}
		if _, duplicate := indices[effect.ClauseID]; duplicate {
			indices[effect.ClauseID] = -1
		} else {
			indices[effect.ClauseID] = index
		}
	}
	normalize := func(reference *compiler.CompiledReference) bool {
		if reference.ProducerClauseID <= 0 ||
			reference.Binding != compiler.ReferenceBindingPriorInstructionResult &&
				reference.Binding != compiler.ReferenceBindingLibraryOwner {
			return true
		}
		index, exists := indices[reference.ProducerClauseID]
		if !exists || index < 0 {
			return false
		}
		reference.PriorInstruction = index
		return true
	}
	normalizeList := func(references []compiler.CompiledReference) ([]compiler.CompiledReference, bool) {
		references = slices.Clone(references)
		for i := range references {
			if !normalize(&references[i]) {
				return nil, false
			}
		}
		return references, true
	}
	var ok bool
	content.References, ok = normalizeList(content.References)
	if !ok {
		return compiler.AbilityContent{}, false
	}
	content.Conditions = slices.Clone(content.Conditions)
	for i := range content.Conditions {
		condition := &content.Conditions[i]
		if condition.ObjectReference == nil || condition.ObjectBinding != compiler.ReferenceBindingPriorInstructionResult {
			continue
		}
		reference := *condition.ObjectReference
		if !normalize(&reference) {
			return compiler.AbilityContent{}, false
		}
		condition.ObjectReference = &reference
	}
	content.Effects = slices.Clone(content.Effects)
	for i := range content.Effects {
		effect := &content.Effects[i]
		effect.References, ok = normalizeList(effect.References)
		if !ok {
			return compiler.AbilityContent{}, false
		}
		effect.SubjectReferences, ok = normalizeList(effect.SubjectReferences)
		if !ok {
			return compiler.AbilityContent{}, false
		}
	}
	return content, true
}
