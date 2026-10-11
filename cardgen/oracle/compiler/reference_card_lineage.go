package compiler

func issueCardLineageSubject(reference *CompiledReference, effects []CompiledEffect) bool {
	if reference.Subject.domain != ReferenceSubjectCard ||
		!reference.LibraryCardObservation ||
		reference.Binding != ReferenceBindingPriorInstructionResult {
		return false
	}
	index := -1
	for i, effect := range effects {
		if effect.ClauseID != reference.ReachedCardProducerClauseID {
			continue
		}
		if index >= 0 || !singularEnteredSubjectProducer(effect, effects) ||
			len(effect.Targets) != 0 || i <= reference.PriorInstruction {
			return false
		}
		index = i
	}
	if index < 0 {
		return false
	}
	inputs := 0
	for _, input := range effects[index].References {
		if input.LibraryCardObservation && input.ProducerClauseID == reference.ProducerClauseID {
			inputs++
		}
	}
	if inputs != 1 {
		return false
	}
	reached := *reference
	reached.ProducerClauseID = reference.ReachedCardProducerClauseID
	reached.ReachedCardProducerClauseID = 0
	reached.LibraryCardObservation = false
	reached.PriorInstruction = index
	reached.Subject = subjectProof(reached, ReferenceSubjectCard, ReferenceLifetimeActualProduct)
	reached.Subject.scope = reference.Subject.scope
	stampSubjectProducer(&reached.Subject, effects[index])
	reached.Subject.enteredProduct = true
	reference.Subject.reached = &reached
	return true
}

// CardLineageSubject supplies the proven entered incarnation of the same card.
// It is an alternative to the still-available input, not an aliased publication.
func (reference CompiledReference) CardLineageSubject(effects []CompiledEffect) (CompiledReference, bool) {
	if !reference.SubjectSupported() || reference.Subject.reached == nil {
		return CompiledReference{}, false
	}
	reached := *reference.Subject.reached
	found := -1
	for i, effect := range effects {
		if effect.ClauseID != reached.ProducerClauseID {
			continue
		}
		if found >= 0 || !reached.SubjectProducerMatches(effect) {
			return CompiledReference{}, false
		}
		found = i
	}
	if found < 0 {
		return CompiledReference{}, false
	}
	reached.PriorInstruction = found
	return reached, true
}
