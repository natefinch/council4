package game

// Kind implements Primitive for Damage.
func (Damage) Kind() PrimitiveKind { return PrimitiveDamage }

// Kind implements Primitive for Draw.
func (Draw) Kind() PrimitiveKind { return PrimitiveDraw }

// Kind implements Primitive for ReorderLibraryTop.
func (ReorderLibraryTop) Kind() PrimitiveKind { return PrimitiveReorderLibraryTop }

// Kind implements Primitive for LookAtLibraryTop.
func (LookAtLibraryTop) Kind() PrimitiveKind { return PrimitiveLookAtLibraryTop }

// Kind implements Primitive for ShuffleLibrary.
func (ShuffleLibrary) Kind() PrimitiveKind { return PrimitiveShuffleLibrary }

// Kind implements Primitive for ShuffleGraveyardIntoLibrary.
func (ShuffleGraveyardIntoLibrary) Kind() PrimitiveKind {
	return PrimitiveShuffleGraveyardIntoLibrary
}

// Kind implements Primitive for LookAtHand.
func (LookAtHand) Kind() PrimitiveKind { return PrimitiveLookAtHand }

// Kind implements Primitive for ChooseDiscardFromHand.
func (ChooseDiscardFromHand) Kind() PrimitiveKind { return PrimitiveChooseDiscardFromHand }

// Kind implements Primitive for ExileEntireHand.
func (ExileEntireHand) Kind() PrimitiveKind { return PrimitiveExileEntireHand }

// Kind implements Primitive for ReturnExiledCardsToHand.
func (ReturnExiledCardsToHand) Kind() PrimitiveKind { return PrimitiveReturnExiledCardsToHand }

// Kind implements Primitive for ExileForEachPlayer.
func (ExileForEachPlayer) Kind() PrimitiveKind { return PrimitiveExileForEachPlayer }

// Kind implements Primitive for ReturnLinkedExiledCardsToBattlefield.
func (ReturnLinkedExiledCardsToBattlefield) Kind() PrimitiveKind {
	return PrimitiveReturnLinkedExiledCardsToBattlefield
}

// Kind implements Primitive for DestroyForEachPlayer.
func (DestroyForEachPlayer) Kind() PrimitiveKind { return PrimitiveDestroyForEachPlayer }

// Kind implements Primitive for CreateTokenForEachDestroyed.
func (CreateTokenForEachDestroyed) Kind() PrimitiveKind {
	return PrimitiveCreateTokenForEachDestroyed
}

// Kind implements Primitive for RemoveTargetsForToken.
func (RemoveTargetsForToken) Kind() PrimitiveKind { return PrimitiveRemoveTargetsForToken }

// Kind implements Primitive for CastForFree.
func (CastForFree) Kind() PrimitiveKind { return PrimitiveCastForFree }

// Kind implements Primitive for ChooseFromZone.
func (ChooseFromZone) Kind() PrimitiveKind { return PrimitiveChooseFromZone }

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

// Kind implements Primitive for TapOrUntap.
func (TapOrUntap) Kind() PrimitiveKind { return PrimitiveTapOrUntap }

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

// Kind implements Primitive for BecomeMonarch.
func (BecomeMonarch) Kind() PrimitiveKind { return PrimitiveBecomeMonarch }

// Kind implements Primitive for SetClassLevel.
func (SetClassLevel) Kind() PrimitiveKind { return PrimitiveSetClassLevel }

// Kind implements Primitive for Monstrosity.
func (Monstrosity) Kind() PrimitiveKind { return PrimitiveMonstrosity }

// Kind implements Primitive for DiscoverCards.
func (DiscoverCards) Kind() PrimitiveKind { return PrimitiveDiscoverCards }

// Kind implements Primitive for Pay.
func (Pay) Kind() PrimitiveKind { return PrimitivePay }

// Kind implements Primitive for PayRepeatedly.
func (PayRepeatedly) Kind() PrimitiveKind { return PrimitivePayRepeatedly }

// Kind implements Primitive for Choose.
func (Choose) Kind() PrimitiveKind { return PrimitiveChoose }

// Kind implements Primitive for GainLife.
func (GainLife) Kind() PrimitiveKind { return PrimitiveGainLife }

// Kind implements Primitive for LoseLife.
func (LoseLife) Kind() PrimitiveKind { return PrimitiveLoseLife }

// Kind implements Primitive for PlayerLosesGame.
func (PlayerLosesGame) Kind() PrimitiveKind { return PrimitivePlayerLosesGame }

// Kind implements Primitive for PlayerWinsGame.
func (PlayerWinsGame) Kind() PrimitiveKind { return PrimitivePlayerWinsGame }

// Kind implements Primitive for PunisherEachLoseLife.
func (PunisherEachLoseLife) Kind() PrimitiveKind { return PrimitivePunisherEachLoseLife }

// Kind implements Primitive for RepeatProcess.
func (RepeatProcess) Kind() PrimitiveKind { return PrimitiveRepeatProcess }

// Kind implements Primitive for CopyStackObject.
func (CopyStackObject) Kind() PrimitiveKind { return PrimitiveCopyStackObject }

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

// Kind implements Primitive for RemoveFromCombat.
func (RemoveFromCombat) Kind() PrimitiveKind { return PrimitiveRemoveFromCombat }

// Kind implements Primitive for CounterObject.
func (CounterObject) Kind() PrimitiveKind { return PrimitiveCounterObject }

// Kind implements Primitive for Mill.
func (Mill) Kind() PrimitiveKind { return PrimitiveMill }

// Kind implements Primitive for ExileTopOfLibrary.
func (ExileTopOfLibrary) Kind() PrimitiveKind { return PrimitiveExileTopOfLibrary }

// Kind implements Primitive for PutHandOnLibraryThenDraw.
func (PutHandOnLibraryThenDraw) Kind() PrimitiveKind { return PrimitivePutHandOnLibraryThenDraw }

// Kind implements Primitive for RevealUntil.
func (RevealUntil) Kind() PrimitiveKind { return PrimitiveRevealUntil }

// Kind implements Primitive for Scry.
func (Scry) Kind() PrimitiveKind { return PrimitiveScry }

// Kind implements Primitive for Surveil.
func (Surveil) Kind() PrimitiveKind { return PrimitiveSurveil }

// Kind implements Primitive for Dig.
func (Dig) Kind() PrimitiveKind { return PrimitiveDig }

// Kind implements Primitive for PileSplit.
func (PileSplit) Kind() PrimitiveKind { return PrimitivePileSplit }

// Kind implements Primitive for ImpulseExile.
func (ImpulseExile) Kind() PrimitiveKind { return PrimitiveImpulseExile }

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

// Kind implements Primitive for BecomeCopy.
func (BecomeCopy) Kind() PrimitiveKind { return PrimitiveBecomeCopy }

// Kind implements Primitive for Amass.
func (Amass) Kind() PrimitiveKind { return PrimitiveAmass }

// Kind implements Primitive for Renown.
func (Renown) Kind() PrimitiveKind { return PrimitiveRenown }

// Kind implements Primitive for Adapt.
func (Adapt) Kind() PrimitiveKind { return PrimitiveAdapt }

// Kind implements Primitive for Connive.
func (Connive) Kind() PrimitiveKind { return PrimitiveConnive }

// Kind implements Primitive for BecomeSaddled.
func (BecomeSaddled) Kind() PrimitiveKind { return PrimitiveBecomeSaddled }

// Kind implements Primitive for ShuffleSpellIntoLibrary.
func (ShuffleSpellIntoLibrary) Kind() PrimitiveKind { return PrimitiveShuffleSpellIntoLibrary }

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

// Kind implements Primitive for MoveCommander.
func (MoveCommander) Kind() PrimitiveKind { return PrimitiveMoveCommander }

// Kind implements Primitive for ChooseNewTargets.
func (ChooseNewTargets) Kind() PrimitiveKind { return PrimitiveChooseNewTargets }

// Kind implements Primitive for GroupSourceDamage.
func (GroupSourceDamage) Kind() PrimitiveKind { return PrimitiveGroupSourceDamage }

// Kind implements Primitive for GroupSelfPowerDamage.
func (GroupSelfPowerDamage) Kind() PrimitiveKind { return PrimitiveGroupSelfPowerDamage }

// Kind implements Primitive for PutPermanentOnLibrary.
func (PutPermanentOnLibrary) Kind() PrimitiveKind { return PrimitivePutPermanentOnLibrary }

// Kind implements Primitive for PutLinkedExiledCardsInLibrary.
func (PutLinkedExiledCardsInLibrary) Kind() PrimitiveKind {
	return PrimitivePutLinkedExiledCardsInLibrary
}

// Kind implements Primitive for GrantCastPermission.
func (GrantCastPermission) Kind() PrimitiveKind { return PrimitiveGrantCastPermission }

// Kind implements Primitive for ExileForPlay.
func (ExileForPlay) Kind() PrimitiveKind { return PrimitiveExileForPlay }

// Kind implements Primitive for HideawayExile.
func (HideawayExile) Kind() PrimitiveKind { return PrimitiveHideawayExile }

// Kind implements Primitive for PlayHideawayCard.
func (PlayHideawayCard) Kind() PrimitiveKind { return PrimitivePlayHideawayCard }

// Kind implements Primitive for Attach.
func (Attach) Kind() PrimitiveKind { return PrimitiveAttach }

// Kind implements Primitive for MassReturnFromGraveyard.
func (MassReturnFromGraveyard) Kind() PrimitiveKind { return PrimitiveMassReturnFromGraveyard }

// Kind implements Primitive for MassReanimationExchange.
func (MassReanimationExchange) Kind() PrimitiveKind { return PrimitiveMassReanimationExchange }

// Kind implements Primitive for AddExtraPhases.
func (AddExtraPhases) Kind() PrimitiveKind { return PrimitiveAddExtraPhases }

// Kind implements Primitive for RollDie.
func (RollDie) Kind() PrimitiveKind { return PrimitiveRollDie }

func (Damage) isPrimitive()                               {}
func (Draw) isPrimitive()                                 {}
func (ReorderLibraryTop) isPrimitive()                    {}
func (LookAtLibraryTop) isPrimitive()                     {}
func (ShuffleLibrary) isPrimitive()                       {}
func (ShuffleGraveyardIntoLibrary) isPrimitive()          {}
func (LookAtHand) isPrimitive()                           {}
func (ChooseDiscardFromHand) isPrimitive()                {}
func (ExileEntireHand) isPrimitive()                      {}
func (ReturnExiledCardsToHand) isPrimitive()              {}
func (ExileForEachPlayer) isPrimitive()                   {}
func (ReturnLinkedExiledCardsToBattlefield) isPrimitive() {}
func (DestroyForEachPlayer) isPrimitive()                 {}
func (CreateTokenForEachDestroyed) isPrimitive()          {}
func (RemoveTargetsForToken) isPrimitive()                {}
func (CastForFree) isPrimitive()                          {}
func (ChooseFromZone) isPrimitive()                       {}
func (Discard) isPrimitive()                              {}
func (Destroy) isPrimitive()                              {}
func (AddMana) isPrimitive()                              {}
func (AddCounter) isPrimitive()                           {}
func (AddPlayerCounter) isPrimitive()                     {}
func (MoveCounters) isPrimitive()                         {}
func (ApplyContinuous) isPrimitive()                      {}
func (ApplyRule) isPrimitive()                            {}
func (ModifyPT) isPrimitive()                             {}
func (Fight) isPrimitive()                                {}
func (Tap) isPrimitive()                                  {}
func (TapOrUntap) isPrimitive()                           {}
func (Search) isPrimitive()                               {}
func (Reveal) isPrimitive()                               {}
func (PutOnBattlefield) isPrimitive()                     {}
func (CreateToken) isPrimitive()                          {}
func (ShufflePermanentIntoLibrary) isPrimitive()          {}
func (StartEngines) isPrimitive()                         {}
func (BecomeMonarch) isPrimitive()                        {}
func (SetClassLevel) isPrimitive()                        {}
func (Monstrosity) isPrimitive()                          {}
func (DiscoverCards) isPrimitive()                        {}
func (Pay) isPrimitive()                                  {}
func (PayRepeatedly) isPrimitive()                        {}
func (Choose) isPrimitive()                               {}
func (GainLife) isPrimitive()                             {}
func (LoseLife) isPrimitive()                             {}
func (PlayerLosesGame) isPrimitive()                      {}
func (PlayerWinsGame) isPrimitive()                       {}
func (PunisherEachLoseLife) isPrimitive()                 {}
func (RepeatProcess) isPrimitive()                        {}
func (CopyStackObject) isPrimitive()                      {}
func (Exile) isPrimitive()                                {}
func (Bounce) isPrimitive()                               {}
func (Sacrifice) isPrimitive()                            {}
func (SacrificePermanents) isPrimitive()                  {}
func (Untap) isPrimitive()                                {}
func (SkipNextUntap) isPrimitive()                        {}
func (RemoveFromCombat) isPrimitive()                     {}
func (CounterObject) isPrimitive()                        {}
func (Mill) isPrimitive()                                 {}
func (ExileTopOfLibrary) isPrimitive()                    {}
func (PutHandOnLibraryThenDraw) isPrimitive()             {}
func (RevealUntil) isPrimitive()                          {}
func (Scry) isPrimitive()                                 {}
func (Surveil) isPrimitive()                              {}
func (Dig) isPrimitive()                                  {}
func (PileSplit) isPrimitive()                            {}
func (ImpulseExile) isPrimitive()                         {}
func (Investigate) isPrimitive()                          {}
func (Proliferate) isPrimitive()                          {}
func (Explore) isPrimitive()                              {}
func (Manifest) isPrimitive()                             {}
func (Goad) isPrimitive()                                 {}
func (RemoveCounter) isPrimitive()                        {}
func (Transform) isPrimitive()                            {}
func (PhaseOut) isPrimitive()                             {}
func (Regenerate) isPrimitive()                           {}
func (BecomeCopy) isPrimitive()                           {}
func (Amass) isPrimitive()                                {}
func (Renown) isPrimitive()                               {}
func (Adapt) isPrimitive()                                {}
func (Connive) isPrimitive()                              {}
func (BecomeSaddled) isPrimitive()                        {}
func (ShuffleSpellIntoLibrary) isPrimitive()              {}
func (SkipStep) isPrimitive()                             {}
func (CreateEmblem) isPrimitive()                         {}
func (CreateDelayedTrigger) isPrimitive()                 {}
func (CreateReplacement) isPrimitive()                    {}
func (PreventDamage) isPrimitive()                        {}
func (MoveCard) isPrimitive()                             {}
func (MoveCommander) isPrimitive()                        {}
func (GrantCastPermission) isPrimitive()                  {}
func (ExileForPlay) isPrimitive()                         {}
func (HideawayExile) isPrimitive()                        {}
func (PlayHideawayCard) isPrimitive()                     {}
func (ChooseNewTargets) isPrimitive()                     {}
func (PutPermanentOnLibrary) isPrimitive()                {}
func (PutLinkedExiledCardsInLibrary) isPrimitive()        {}
func (Attach) isPrimitive()                               {}
func (MassReturnFromGraveyard) isPrimitive()              {}

func (GroupSourceDamage) isPrimitive() {}

func (GroupSelfPowerDamage) isPrimitive() {}

func (MassReanimationExchange) isPrimitive() {}

func (AddExtraPhases) isPrimitive() {}

func (RollDie) isPrimitive() {}

func (p Damage) instructionRefs() primitiveRefs { return quantityRefs(p.Amount) }
func (p Draw) instructionRefs() primitiveRefs   { return quantityRefs(p.Amount) }
func (p ReorderLibraryTop) instructionRefs() primitiveRefs {
	return quantityRefs(p.Amount)
}
func (p LookAtLibraryTop) instructionRefs() primitiveRefs {
	return primitiveRefs{publishesLinked: p.PublishLinked}
}
func (ShuffleLibrary) instructionRefs() primitiveRefs { return primitiveRefs{} }
func (ShuffleGraveyardIntoLibrary) instructionRefs() primitiveRefs {
	return primitiveRefs{}
}
func (LookAtHand) instructionRefs() primitiveRefs            { return primitiveRefs{} }
func (ChooseDiscardFromHand) instructionRefs() primitiveRefs { return primitiveRefs{} }
func (p Discard) instructionRefs() primitiveRefs {
	refs := quantityRefs(p.Amount)
	refs.publishesLinked = p.PublishLinked
	return refs
}
func (Destroy) instructionRefs() primitiveRefs      { return primitiveRefs{} }
func (p AddCounter) instructionRefs() primitiveRefs { return quantityRefs(p.Amount) }
func (p AddPlayerCounter) instructionRefs() primitiveRefs {
	return quantityRefs(p.Amount)
}
func (p MoveCounters) instructionRefs() primitiveRefs { return quantityRefs(p.Amount) }
func (p ApplyContinuous) instructionRefs() primitiveRefs {
	refs := quantityRefs(p.ChooseUpTo)
	refs.publishesLinked = p.PublishLinked
	return refs
}
func (ApplyRule) instructionRefs() primitiveRefs { return primitiveRefs{} }

func (p ModifyPT) instructionRefs() primitiveRefs {
	refs := mergePrimitiveRefs(objectReferenceRefs(p.Object), quantityRefs(p.PowerDelta))
	refs = mergePrimitiveRefs(refs, quantityRefs(p.ToughnessDelta))
	refs.publishesLinked = p.PublishLinked
	return refs
}
func (Fight) instructionRefs() primitiveRefs      { return primitiveRefs{} }
func (Tap) instructionRefs() primitiveRefs        { return primitiveRefs{} }
func (TapOrUntap) instructionRefs() primitiveRefs { return primitiveRefs{} }
func (p Search) instructionRefs() primitiveRefs {
	refs := quantityRefs(p.Amount)
	refs.publishesLinked = p.PublishLinked
	return refs
}

func (p CreateToken) instructionRefs() primitiveRefs {
	refs := quantityRefs(p.Amount)
	refs.publishesLinked = p.PublishLinked
	return refs
}
func (ShufflePermanentIntoLibrary) instructionRefs() primitiveRefs { return primitiveRefs{} }
func (StartEngines) instructionRefs() primitiveRefs                { return primitiveRefs{} }
func (BecomeMonarch) instructionRefs() primitiveRefs               { return primitiveRefs{} }
func (p SetClassLevel) instructionRefs() primitiveRefs             { return quantityRefs(p.Amount) }
func (p Monstrosity) instructionRefs() primitiveRefs               { return quantityRefs(p.Amount) }
func (p DiscoverCards) instructionRefs() primitiveRefs             { return quantityRefs(p.Amount) }
func (Pay) instructionRefs() primitiveRefs                         { return primitiveRefs{} }

func (p PayRepeatedly) instructionRefs() primitiveRefs {
	return primitiveRefs{publishesChoice: ChoiceKey(p.PublishCount)}
}

func (p AddMana) instructionRefs() primitiveRefs {
	refs := quantityRefs(p.Amount)
	if p.ChoiceFrom != "" {
		refs.consumesChoices = append(refs.consumesChoices, p.ChoiceFrom)
	}
	return refs
}

func (p Reveal) instructionRefs() primitiveRefs {
	if p.Card.Kind != CardReferenceNone {
		return cardReferenceRefs(p.Card)
	}
	refs := quantityRefs(p.Amount)
	refs.publishesLinked = p.PublishLinked
	return refs
}

func (p PutOnBattlefield) instructionRefs() primitiveRefs {
	refs := primitiveRefs{publishesLinked: p.PublishLinked}
	if key := p.Source.sourceLinkedKey(); key != "" {
		refs.consumesLinked = []LinkedKey{key}
	}
	for _, source := range p.Sources {
		if key := source.sourceLinkedKey(); key != "" {
			refs.consumesLinked = append(refs.consumesLinked, key)
		}
	}
	return refs
}

func (p Choose) instructionRefs() primitiveRefs {
	return primitiveRefs{publishesChoice: p.PublishChoice}
}

func (p GainLife) instructionRefs() primitiveRefs             { return quantityRefs(p.Amount) }
func (p LoseLife) instructionRefs() primitiveRefs             { return quantityRefs(p.Amount) }
func (PlayerLosesGame) instructionRefs() primitiveRefs        { return primitiveRefs{} }
func (PlayerWinsGame) instructionRefs() primitiveRefs         { return primitiveRefs{} }
func (p PunisherEachLoseLife) instructionRefs() primitiveRefs { return quantityRefs(p.Amount) }
func (p RepeatProcess) instructionRefs() primitiveRefs        { return quantityRefs(p.Times) }

func (p Exile) instructionRefs() primitiveRefs {
	return primitiveRefs{publishesLinked: p.ExileLinkedKey}
}

func (p ExileEntireHand) instructionRefs() primitiveRefs {
	return primitiveRefs{publishesLinked: p.LinkedKey}
}
func (p ReturnExiledCardsToHand) instructionRefs() primitiveRefs {
	return primitiveRefs{consumesLinked: []LinkedKey{p.LinkedKey}}
}
func (p ExileForEachPlayer) instructionRefs() primitiveRefs {
	return primitiveRefs{publishesLinked: p.LinkedKey}
}
func (p DestroyForEachPlayer) instructionRefs() primitiveRefs {
	return primitiveRefs{publishesLinked: p.LinkedKey}
}
func (p CreateTokenForEachDestroyed) instructionRefs() primitiveRefs {
	return primitiveRefs{consumesLinked: []LinkedKey{p.LinkedKey}}
}
func (p RemoveTargetsForToken) instructionRefs() primitiveRefs {
	return primitiveRefs{publishesLinked: p.LinkedKey}
}
func (p ReturnLinkedExiledCardsToBattlefield) instructionRefs() primitiveRefs {
	refs := quantityRefs(p.Amount)
	refs.consumesLinked = append(refs.consumesLinked, p.LinkedKey)
	return refs
}
func (CastForFree) instructionRefs() primitiveRefs { return primitiveRefs{} }
func (p ChooseFromZone) instructionRefs() primitiveRefs {
	refs := quantityRefs(p.Quantity)
	if p.Riders.FromLinked != "" {
		refs.consumesLinked = append(refs.consumesLinked, p.Riders.FromLinked)
	}
	refs.publishesLinked = p.Riders.PublishLinked
	return refs
}
func (p Bounce) instructionRefs() primitiveRefs  { return objectReferenceRefs(p.Object) }
func (Sacrifice) instructionRefs() primitiveRefs { return primitiveRefs{} }
func (p SacrificePermanents) instructionRefs() primitiveRefs {
	refs := quantityRefs(p.Amount)
	if p.PublishLinked != "" {
		refs.publishesLinked = p.PublishLinked
	}
	return refs
}
func (p Untap) instructionRefs() primitiveRefs {
	return mergePrimitiveRefs(objectReferenceRefs(p.Object), quantityRefs(p.Amount))
}
func (SkipNextUntap) instructionRefs() primitiveRefs    { return primitiveRefs{} }
func (RemoveFromCombat) instructionRefs() primitiveRefs { return primitiveRefs{} }
func (CounterObject) instructionRefs() primitiveRefs    { return primitiveRefs{} }
func (p Mill) instructionRefs() primitiveRefs {
	refs := quantityRefs(p.Amount)
	refs.publishesLinked = p.PublishLinked
	return refs
}
func (p ExileTopOfLibrary) instructionRefs() primitiveRefs {
	return quantityRefs(p.Amount)
}
func (PutHandOnLibraryThenDraw) instructionRefs() primitiveRefs { return primitiveRefs{} }
func (RevealUntil) instructionRefs() primitiveRefs              { return primitiveRefs{} }
func (p Scry) instructionRefs() primitiveRefs                   { return quantityRefs(p.Amount) }
func (p Surveil) instructionRefs() primitiveRefs                { return quantityRefs(p.Amount) }
func (p Dig) instructionRefs() primitiveRefs                    { return quantityRefs(p.Look) }
func (p PileSplit) instructionRefs() primitiveRefs              { return quantityRefs(p.Amount) }
func (p ImpulseExile) instructionRefs() primitiveRefs           { return quantityRefs(p.Amount) }
func (p Investigate) instructionRefs() primitiveRefs            { return quantityRefs(p.Amount) }
func (p Proliferate) instructionRefs() primitiveRefs            { return quantityRefs(p.Amount) }
func (Explore) instructionRefs() primitiveRefs                  { return primitiveRefs{} }
func (p Manifest) instructionRefs() primitiveRefs {
	return primitiveRefs{publishesLinked: p.PublishLinked}
}
func (Goad) instructionRefs() primitiveRefs { return primitiveRefs{} }

func (p RemoveCounter) instructionRefs() primitiveRefs         { return quantityRefs(p.Amount) }
func (Transform) instructionRefs() primitiveRefs               { return primitiveRefs{} }
func (PhaseOut) instructionRefs() primitiveRefs                { return primitiveRefs{} }
func (Regenerate) instructionRefs() primitiveRefs              { return primitiveRefs{} }
func (BecomeCopy) instructionRefs() primitiveRefs              { return primitiveRefs{} }
func (p Amass) instructionRefs() primitiveRefs                 { return quantityRefs(p.Amount) }
func (p Renown) instructionRefs() primitiveRefs                { return quantityRefs(p.Amount) }
func (p Adapt) instructionRefs() primitiveRefs                 { return quantityRefs(p.Amount) }
func (p Connive) instructionRefs() primitiveRefs               { return quantityRefs(p.Amount) }
func (BecomeSaddled) instructionRefs() primitiveRefs           { return primitiveRefs{} }
func (ShuffleSpellIntoLibrary) instructionRefs() primitiveRefs { return primitiveRefs{} }
func (SkipStep) instructionRefs() primitiveRefs                { return primitiveRefs{} }
func (CreateEmblem) instructionRefs() primitiveRefs            { return primitiveRefs{} }
func (p CreateDelayedTrigger) instructionRefs() primitiveRefs {
	if !p.Trigger.DamageSourceObject.Exists {
		return primitiveRefs{}
	}
	return objectReferenceRefs(p.Trigger.DamageSourceObject.Val)
}
func (p CreateReplacement) instructionRefs() primitiveRefs { return objectReferenceRefs(p.Object) }
func (p PreventDamage) instructionRefs() primitiveRefs     { return quantityRefs(p.Amount) }
func (p MoveCard) instructionRefs() primitiveRefs {
	if p.Player.Kind() != PlayerReferenceNone {
		return quantityRefs(p.Amount)
	}
	return cardReferenceRefs(p.Card)
}
func (MoveCommander) instructionRefs() primitiveRefs        { return primitiveRefs{} }
func (ChooseNewTargets) instructionRefs() primitiveRefs     { return primitiveRefs{} }
func (CopyStackObject) instructionRefs() primitiveRefs      { return primitiveRefs{} }
func (p GroupSourceDamage) instructionRefs() primitiveRefs  { return quantityRefs(p.Amount) }
func (GroupSelfPowerDamage) instructionRefs() primitiveRefs { return primitiveRefs{} }
func (MassReturnFromGraveyard) instructionRefs() primitiveRefs {
	return primitiveRefs{}
}
func (MassReanimationExchange) instructionRefs() primitiveRefs {
	return primitiveRefs{}
}
func (AddExtraPhases) instructionRefs() primitiveRefs {
	return primitiveRefs{}
}
func (RollDie) instructionRefs() primitiveRefs {
	return primitiveRefs{}
}
func (p GrantCastPermission) instructionRefs() primitiveRefs {
	return cardReferenceRefs(p.Card)
}
func (p ExileForPlay) instructionRefs() primitiveRefs {
	return cardReferenceRefs(p.Card)
}
func (p HideawayExile) instructionRefs() primitiveRefs {
	return quantityRefs(p.Amount)
}
func (PlayHideawayCard) instructionRefs() primitiveRefs { return primitiveRefs{} }
func (p PutPermanentOnLibrary) instructionRefs() primitiveRefs {
	return objectReferenceRefs(p.Object)
}

func (p PutLinkedExiledCardsInLibrary) instructionRefs() primitiveRefs {
	if p.LinkedKey == "" {
		return primitiveRefs{}
	}
	return primitiveRefs{consumesLinked: []LinkedKey{p.LinkedKey}}
}

func (p Attach) instructionRefs() primitiveRefs {
	return mergePrimitiveRefs(objectReferenceRefs(p.Attachment), objectReferenceRefs(p.Target))
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
	refs := objectReferenceRefs(dynamic.Object)
	switch dynamic.Kind {
	case DynamicAmountPreviousEffectResult, DynamicAmountPreviousEffectExcessDamage:
		if dynamic.ResultKey != "" {
			refs.consumesResults = append(refs.consumesResults, dynamic.ResultKey)
		}
	case DynamicAmountChosenNumber:
		if dynamic.ResultKey != "" {
			refs.consumesChoices = append(refs.consumesChoices, ChoiceKey(dynamic.ResultKey))
		}
	case DynamicAmountMaxOf:
		for i := range dynamic.Operands {
			refs = mergePrimitiveRefs(refs, quantityRefs(Dynamic(dynamic.Operands[i])))
		}
	default:
	}
	return refs
}

func mergePrimitiveRefs(left, right primitiveRefs) primitiveRefs {
	left.consumesResults = append(left.consumesResults, right.consumesResults...)
	left.consumesChoices = append(left.consumesChoices, right.consumesChoices...)
	left.consumesLinked = append(left.consumesLinked, right.consumesLinked...)
	return left
}
