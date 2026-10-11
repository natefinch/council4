package compiler

import "github.com/natefinch/council4/cardgen/oracle/shared"

// TargetActionSubject resolves one clause's declared or inherited target without
// changing its canonical subject identity when assembly localizes target slots.
func (content AbilityContent) TargetActionSubject(effect CompiledEffect) (CompiledReference, bool) {
	if content.subjectScope == nil || content.subjectScope.owner == nil ||
		effect.subjectScope != content.subjectScope || effect.ClauseID <= 0 ||
		len(content.Targets) != 1 || len(content.References) > 1 {
		return CompiledReference{}, false
	}
	owner := content.subjectScope.owner
	var original *CompiledEffect
	for i := range owner.Effects {
		candidate := &owner.Effects[i]
		if candidate.ClauseID == effect.ClauseID {
			if original != nil {
				return CompiledReference{}, false
			}
			original = candidate
		}
	}
	if original == nil || original.Kind != effect.Kind || original.Order != effect.Order ||
		effect.Order == (shared.SourceOrder{}) {
		return CompiledReference{}, false
	}
	target := content.Targets[0]
	if !target.Exact || target.Cardinality.Min != 1 || target.Cardinality.Max != 1 {
		return CompiledReference{}, false
	}
	if len(content.References) == 0 {
		if len(effect.Targets) != 1 || len(original.Targets) != 1 ||
			effect.Targets[0].Order != target.Order || original.Targets[0].Order != target.Order {
			return CompiledReference{}, false
		}
		return content.TargetSubject(0)
	}
	if len(effect.Targets) != 0 || len(original.Targets) != 0 {
		return CompiledReference{}, false
	}
	reference := content.References[0]
	if !content.OwnsSubject(reference) ||
		reference.Binding != ReferenceBindingTarget || !effectOwnsReference(*original, reference) {
		return CompiledReference{}, false
	}
	var canonical *CompiledReference
	for i := range owner.References {
		candidate := &owner.References[i]
		if candidate.NodeID == reference.NodeID {
			if canonical != nil {
				return CompiledReference{}, false
			}
			canonical = candidate
		}
	}
	if canonical == nil || canonical.Order != reference.Order ||
		target.Order.End > reference.Order.Start ||
		!original.Order.Contains(reference.Order) {
		return CompiledReference{}, false
	}
	localized, ok := canonical.LocalizeSubjectOccurrence(target, 0)
	if !ok || localized.Subject != reference.Subject || localized.Occurrence != reference.Occurrence {
		return CompiledReference{}, false
	}
	return reference, true
}
