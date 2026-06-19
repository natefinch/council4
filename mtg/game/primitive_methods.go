package game

// Kind implements Primitive for Damage.
func (Damage) Kind() PrimitiveKind { return PrimitiveDamage }

// Kind implements Primitive for Draw.
func (Draw) Kind() PrimitiveKind { return PrimitiveDraw }

// Kind implements Primitive for Discard.
func (Discard) Kind() PrimitiveKind { return PrimitiveDiscard }

// Kind implements Primitive for Destroy.
func (Destroy) Kind() PrimitiveKind { return PrimitiveDestroy }

// Kind implements Primitive for AddMana.
func (AddMana) Kind() PrimitiveKind { return PrimitiveAddMana }

// Kind implements Primitive for AddCounter.
func (AddCounter) Kind() PrimitiveKind { return PrimitiveAddCounter }

// Kind implements Primitive for AddPlayerCounter.
func (AddPlayerCounter) Kind() PrimitiveKind { return PrimitiveAddPlayerCounter }

// Kind implements Primitive for MoveCounters.
func (MoveCounters) Kind() PrimitiveKind { return PrimitiveMoveCounters }

// Kind implements Primitive for ApplyContinuous.
func (ApplyContinuous) Kind() PrimitiveKind { return PrimitiveApplyContinuous }

// Kind implements Primitive for ApplyRule.
func (ApplyRule) Kind() PrimitiveKind { return PrimitiveApplyRule }

// Kind implements Primitive for ModifyPT.
func (ModifyPT) Kind() PrimitiveKind { return PrimitiveModifyPT }

// Kind implements Primitive for Fight.
func (Fight) Kind() PrimitiveKind { return PrimitiveFight }

// Kind implements Primitive for Tap.
func (Tap) Kind() PrimitiveKind { return PrimitiveTap }

// Kind implements Primitive for Search.
func (Search) Kind() PrimitiveKind { return PrimitiveSearch }

// Kind implements Primitive for Reveal.
func (Reveal) Kind() PrimitiveKind { return PrimitiveReveal }

// Kind implements Primitive for PutOnBattlefield.
func (PutOnBattlefield) Kind() PrimitiveKind { return PrimitivePutOnBattlefield }

// Kind implements Primitive for CreateToken.
func (CreateToken) Kind() PrimitiveKind { return PrimitiveCreateToken }

// Kind implements Primitive for ShufflePermanentIntoLibrary.
func (ShufflePermanentIntoLibrary) Kind() PrimitiveKind { return PrimitiveShufflePermanentIntoLibrary }

// Kind implements Primitive for StartEngines.
func (StartEngines) Kind() PrimitiveKind { return PrimitiveStartEngines }

// Kind implements Primitive for SetClassLevel.
func (SetClassLevel) Kind() PrimitiveKind { return PrimitiveSetClassLevel }

// Kind implements Primitive for Monstrosity.
func (Monstrosity) Kind() PrimitiveKind { return PrimitiveMonstrosity }

// Kind implements Primitive for DiscoverCards.
func (DiscoverCards) Kind() PrimitiveKind { return PrimitiveDiscoverCards }

// Kind implements Primitive for Pay.
func (Pay) Kind() PrimitiveKind { return PrimitivePay }

// Kind implements Primitive for Choose.
func (Choose) Kind() PrimitiveKind { return PrimitiveChoose }

// Kind implements Primitive for GainLife.
func (GainLife) Kind() PrimitiveKind { return PrimitiveGainLife }

// Kind implements Primitive for LoseLife.
func (LoseLife) Kind() PrimitiveKind { return PrimitiveLoseLife }

// Kind implements Primitive for Exile.
func (Exile) Kind() PrimitiveKind { return PrimitiveExile }

// Kind implements Primitive for Bounce.
func (Bounce) Kind() PrimitiveKind { return PrimitiveBounce }

// Kind implements Primitive for Sacrifice.
func (Sacrifice) Kind() PrimitiveKind { return PrimitiveSacrifice }

// Kind implements Primitive for SacrificePermanents.
func (SacrificePermanents) Kind() PrimitiveKind { return PrimitiveSacrificePermanents }

// Kind implements Primitive for Untap.
func (Untap) Kind() PrimitiveKind { return PrimitiveUntap }

// Kind implements Primitive for SkipNextUntap.
func (SkipNextUntap) Kind() PrimitiveKind { return PrimitiveSkipNextUntap }

// Kind implements Primitive for CounterObject.
func (CounterObject) Kind() PrimitiveKind { return PrimitiveCounterObject }

// Kind implements Primitive for Mill.
func (Mill) Kind() PrimitiveKind { return PrimitiveMill }

// Kind implements Primitive for Scry.
func (Scry) Kind() PrimitiveKind { return PrimitiveScry }

// Kind implements Primitive for Surveil.
func (Surveil) Kind() PrimitiveKind { return PrimitiveSurveil }

// Kind implements Primitive for Dig.
func (Dig) Kind() PrimitiveKind { return PrimitiveDig }

// Kind implements Primitive for Investigate.
func (Investigate) Kind() PrimitiveKind { return PrimitiveInvestigate }

// Kind implements Primitive for Proliferate.
func (Proliferate) Kind() PrimitiveKind { return PrimitiveProliferate }

// Kind implements Primitive for Explore.
func (Explore) Kind() PrimitiveKind { return PrimitiveExplore }

// Kind implements Primitive for Manifest.
func (Manifest) Kind() PrimitiveKind { return PrimitiveManifest }

// Kind implements Primitive for Goad.
func (Goad) Kind() PrimitiveKind { return PrimitiveGoad }

// Kind implements Primitive for RemoveCounter.
func (RemoveCounter) Kind() PrimitiveKind { return PrimitiveRemoveCounter }

// Kind implements Primitive for Transform.
func (Transform) Kind() PrimitiveKind { return PrimitiveTransform }

// Kind implements Primitive for PhaseOut.
func (PhaseOut) Kind() PrimitiveKind { return PrimitivePhaseOut }

// Kind implements Primitive for Regenerate.
func (Regenerate) Kind() PrimitiveKind { return PrimitiveRegenerate }

// Kind implements Primitive for SkipStep.
func (SkipStep) Kind() PrimitiveKind { return PrimitiveSkipStep }

// Kind implements Primitive for CreateEmblem.
func (CreateEmblem) Kind() PrimitiveKind { return PrimitiveCreateEmblem }

// Kind implements Primitive for CreateDelayedTrigger.
func (CreateDelayedTrigger) Kind() PrimitiveKind { return PrimitiveCreateDelayedTrigger }

// Kind implements Primitive for CreateReplacement.
func (CreateReplacement) Kind() PrimitiveKind { return PrimitiveCreateReplacement }

// Kind implements Primitive for PreventDamage.
func (PreventDamage) Kind() PrimitiveKind { return PrimitivePreventDamage }

// Kind implements Primitive for MoveCard.
func (MoveCard) Kind() PrimitiveKind { return PrimitiveMoveCard }

// Kind implements Primitive for GrantCastPermission.
func (GrantCastPermission) Kind() PrimitiveKind { return PrimitiveGrantCastPermission }

func (Damage) isPrimitive()                      {}
func (Draw) isPrimitive()                        {}
func (Discard) isPrimitive()                     {}
func (Destroy) isPrimitive()                     {}
func (AddMana) isPrimitive()                     {}
func (AddCounter) isPrimitive()                  {}
func (AddPlayerCounter) isPrimitive()            {}
func (MoveCounters) isPrimitive()                {}
func (ApplyContinuous) isPrimitive()             {}
func (ApplyRule) isPrimitive()                   {}
func (ModifyPT) isPrimitive()                    {}
func (Fight) isPrimitive()                       {}
func (Tap) isPrimitive()                         {}
func (Search) isPrimitive()                      {}
func (Reveal) isPrimitive()                      {}
func (PutOnBattlefield) isPrimitive()            {}
func (CreateToken) isPrimitive()                 {}
func (ShufflePermanentIntoLibrary) isPrimitive() {}
func (StartEngines) isPrimitive()                {}
func (SetClassLevel) isPrimitive()               {}
func (Monstrosity) isPrimitive()                 {}
func (DiscoverCards) isPrimitive()               {}
func (Pay) isPrimitive()                         {}
func (Choose) isPrimitive()                      {}
func (GainLife) isPrimitive()                    {}
func (LoseLife) isPrimitive()                    {}
func (Exile) isPrimitive()                       {}
func (Bounce) isPrimitive()                      {}
func (Sacrifice) isPrimitive()                   {}
func (SacrificePermanents) isPrimitive()         {}
func (Untap) isPrimitive()                       {}
func (SkipNextUntap) isPrimitive()               {}
func (CounterObject) isPrimitive()               {}
func (Mill) isPrimitive()                        {}
func (Scry) isPrimitive()                        {}
func (Surveil) isPrimitive()                     {}
func (Dig) isPrimitive()                         {}
func (Investigate) isPrimitive()                 {}
func (Proliferate) isPrimitive()                 {}
func (Explore) isPrimitive()                     {}
func (Manifest) isPrimitive()                    {}
func (Goad) isPrimitive()                        {}
func (RemoveCounter) isPrimitive()               {}
func (Transform) isPrimitive()                   {}
func (PhaseOut) isPrimitive()                    {}
func (Regenerate) isPrimitive()                  {}
func (SkipStep) isPrimitive()                    {}
func (CreateEmblem) isPrimitive()                {}
func (CreateDelayedTrigger) isPrimitive()        {}
func (CreateReplacement) isPrimitive()           {}
func (PreventDamage) isPrimitive()               {}
func (MoveCard) isPrimitive()                    {}
func (GrantCastPermission) isPrimitive()         {}

func (p Damage) instructionRefs() primitiveRefs     { return quantityRefs(p.Amount) }
func (p Draw) instructionRefs() primitiveRefs       { return quantityRefs(p.Amount) }
func (p Discard) instructionRefs() primitiveRefs    { return quantityRefs(p.Amount) }
func (Destroy) instructionRefs() primitiveRefs      { return primitiveRefs{} }
func (p AddCounter) instructionRefs() primitiveRefs { return quantityRefs(p.Amount) }
func (p AddPlayerCounter) instructionRefs() primitiveRefs {
	return quantityRefs(p.Amount)
}
func (p MoveCounters) instructionRefs() primitiveRefs { return quantityRefs(p.Amount) }
func (p ApplyContinuous) instructionRefs() primitiveRefs {
	return primitiveRefs{publishesLinked: p.PublishLinked}
}
func (ApplyRule) instructionRefs() primitiveRefs { return primitiveRefs{} }

func (p ModifyPT) instructionRefs() primitiveRefs {
	refs := mergePrimitiveRefs(objectReferenceRefs(p.Object), quantityRefs(p.PowerDelta))
	refs = mergePrimitiveRefs(refs, quantityRefs(p.ToughnessDelta))
	refs.publishesLinked = p.PublishLinked
	return refs
}
func (Fight) instructionRefs() primitiveRefs { return primitiveRefs{} }
func (Tap) instructionRefs() primitiveRefs   { return primitiveRefs{} }
func (p Search) instructionRefs() primitiveRefs {
	refs := quantityRefs(p.Amount)
	refs.publishesLinked = p.PublishLinked
	return refs
}

func (p CreateToken) instructionRefs() primitiveRefs               { return quantityRefs(p.Amount) }
func (ShufflePermanentIntoLibrary) instructionRefs() primitiveRefs { return primitiveRefs{} }
func (StartEngines) instructionRefs() primitiveRefs                { return primitiveRefs{} }
func (p SetClassLevel) instructionRefs() primitiveRefs             { return quantityRefs(p.Amount) }
func (p Monstrosity) instructionRefs() primitiveRefs               { return quantityRefs(p.Amount) }
func (p DiscoverCards) instructionRefs() primitiveRefs             { return quantityRefs(p.Amount) }
func (Pay) instructionRefs() primitiveRefs                         { return primitiveRefs{} }

func (p AddMana) instructionRefs() primitiveRefs {
	refs := quantityRefs(p.Amount)
	if p.ChoiceFrom != "" {
		refs.consumesChoices = append(refs.consumesChoices, p.ChoiceFrom)
	}
	return refs
}

func (p Reveal) instructionRefs() primitiveRefs {
	refs := quantityRefs(p.Amount)
	refs.publishesLinked = p.PublishLinked
	return refs
}

func (p PutOnBattlefield) instructionRefs() primitiveRefs {
	refs := primitiveRefs{publishesLinked: p.PublishLinked}
	if key := p.Source.sourceLinkedKey(); key != "" {
		refs.consumesLinked = []LinkedKey{key}
	}
	return refs
}

func (p Choose) instructionRefs() primitiveRefs {
	return primitiveRefs{publishesChoice: p.PublishChoice}
}

func (p GainLife) instructionRefs() primitiveRefs { return quantityRefs(p.Amount) }
func (p LoseLife) instructionRefs() primitiveRefs { return quantityRefs(p.Amount) }

func (p Exile) instructionRefs() primitiveRefs {
	return primitiveRefs{publishesLinked: p.ExileLinkedKey}
}
func (p Bounce) instructionRefs() primitiveRefs              { return objectReferenceRefs(p.Object) }
func (Sacrifice) instructionRefs() primitiveRefs             { return primitiveRefs{} }
func (p SacrificePermanents) instructionRefs() primitiveRefs { return quantityRefs(p.Amount) }
func (p Untap) instructionRefs() primitiveRefs               { return objectReferenceRefs(p.Object) }
func (SkipNextUntap) instructionRefs() primitiveRefs         { return primitiveRefs{} }
func (CounterObject) instructionRefs() primitiveRefs         { return primitiveRefs{} }
func (p Mill) instructionRefs() primitiveRefs                { return quantityRefs(p.Amount) }
func (p Scry) instructionRefs() primitiveRefs                { return quantityRefs(p.Amount) }
func (p Surveil) instructionRefs() primitiveRefs             { return quantityRefs(p.Amount) }
func (p Dig) instructionRefs() primitiveRefs                 { return quantityRefs(p.Look) }
func (p Investigate) instructionRefs() primitiveRefs         { return quantityRefs(p.Amount) }
func (p Proliferate) instructionRefs() primitiveRefs         { return quantityRefs(p.Amount) }
func (Explore) instructionRefs() primitiveRefs               { return primitiveRefs{} }
func (Manifest) instructionRefs() primitiveRefs              { return primitiveRefs{} }
func (Goad) instructionRefs() primitiveRefs                  { return primitiveRefs{} }

func (p RemoveCounter) instructionRefs() primitiveRefs      { return quantityRefs(p.Amount) }
func (Transform) instructionRefs() primitiveRefs            { return primitiveRefs{} }
func (PhaseOut) instructionRefs() primitiveRefs             { return primitiveRefs{} }
func (Regenerate) instructionRefs() primitiveRefs           { return primitiveRefs{} }
func (SkipStep) instructionRefs() primitiveRefs             { return primitiveRefs{} }
func (CreateEmblem) instructionRefs() primitiveRefs         { return primitiveRefs{} }
func (CreateDelayedTrigger) instructionRefs() primitiveRefs { return primitiveRefs{} }
func (CreateReplacement) instructionRefs() primitiveRefs    { return primitiveRefs{} }
func (p PreventDamage) instructionRefs() primitiveRefs      { return quantityRefs(p.Amount) }
func (p MoveCard) instructionRefs() primitiveRefs           { return cardReferenceRefs(p.Card) }
func (p GrantCastPermission) instructionRefs() primitiveRefs {
	return cardReferenceRefs(p.Card)
}

func cardReferenceRefs(reference CardReference) primitiveRefs {
	if reference.Kind != CardReferenceLinked || reference.LinkID == "" {
		return primitiveRefs{}
	}
	return primitiveRefs{consumesLinked: []LinkedKey{LinkedKey(reference.LinkID)}}
}

func objectReferenceRefs(reference ObjectReference) primitiveRefs {
	if reference.Kind() != ObjectReferenceLinkedObject || reference.LinkID() == "" {
		return primitiveRefs{}
	}
	return primitiveRefs{consumesLinked: []LinkedKey{LinkedKey(reference.LinkID())}}
}

func quantityRefs(quantity Quantity) primitiveRefs {
	if !quantity.IsDynamic() {
		return primitiveRefs{}
	}
	dynamic := quantity.DynamicAmount().Val
	switch dynamic.Kind {
	case DynamicAmountPreviousEffectResult, DynamicAmountPreviousEffectExcessDamage:
		if dynamic.ResultKey != "" {
			return primitiveRefs{consumesResults: []ResultKey{dynamic.ResultKey}}
		}
	case DynamicAmountChosenNumber:
		if dynamic.ResultKey != "" {
			return primitiveRefs{consumesChoices: []ChoiceKey{ChoiceKey(dynamic.ResultKey)}}
		}
	default:
	}
	return primitiveRefs{}
}

func mergePrimitiveRefs(left, right primitiveRefs) primitiveRefs {
	left.consumesResults = append(left.consumesResults, right.consumesResults...)
	left.consumesChoices = append(left.consumesChoices, right.consumesChoices...)
	left.consumesLinked = append(left.consumesLinked, right.consumesLinked...)
	return left
}
