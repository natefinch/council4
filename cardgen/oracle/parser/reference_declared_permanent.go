package parser

import "github.com/natefinch/council4/mtg/game/zone"

func emitDeclaredPermanentSubjects(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		bindDeclaredPermanentSubjects(ability.Sentences, ability.SemanticReferences)
		if ability.Modal != nil {
			for j := range ability.Modal.Options {
				mode := &ability.Modal.Options[j]
				bindDeclaredPermanentSubjects(mode.Sentences, mode.SemanticReferences)
			}
		}
	}
}

func bindDeclaredPermanentSubjects(sentences []Sentence, references []Reference) {
	var effects []*EffectSyntax
	for i := range sentences {
		for j := range sentences[i].Effects {
			effects = append(effects, &sentences[i].Effects[j])
		}
	}
	for i := range references {
		reference := &references[i]
		if reference.ProducerClauseID != 0 || reference.LibraryCardObservation || reference.CardIdentity ||
			!declaredPermanentReference(*reference) || !declaredPermanentProjection(*reference, effects) {
			continue
		}
		var subject *EffectSyntax
		permanents := 0
		cardResult := false
		for _, effect := range effects {
			if effect.ClauseSpan.End.Offset > reference.Span.Start.Offset {
				continue
			}
			if len(effect.Targets) != 0 || enteredSubjectBarrier(effect) ||
				effect.Kind == EffectCounter || effect.Kind == EffectBolster ||
				effect.Kind == EffectManifest || effect.Kind == EffectCloak || effect.Kind == EffectPopulate {
				subject = effect
			}
			for _, target := range effect.Targets {
				if declaredPermanentSelection(target.Selection) {
					permanents++
				} else if effect.Kind == EffectExile && target.Selection.Zone != zone.None &&
					target.Selection.Zone != zone.Battlefield {
					cardResult = true
				}
			}
		}
		if subject == nil || len(subject.Targets) == 0 || closerSourceAntecedent(*reference, subject, references) {
			continue
		}
		permanent := false
		for _, target := range subject.Targets {
			permanent = permanent || declaredPermanentSelection(target.Selection)
		}
		if !permanent {
			continue
		}
		if len(subject.Targets) == 1 && !enteredNounCompatible(reference.SubjectNoun, subject.Targets[0].Selection.Kind) ||
			!declaredPermanentAction(subject) {
			continue
		}
		if !cardResult && (permanents != 1 || len(subject.Targets) != 1 ||
			subject.Targets[0].Cardinality.Min != 1 || subject.Targets[0].Cardinality.Max != 1) {
			continue
		}
		reference.DeclaredPermanent = &DeclaredPermanentSubject{}
		if permanents != 1 || len(subject.Targets) != 1 || !subject.Exact || subject.Negated ||
			subject.DelayedTiming != DelayedTimingNone {
			continue
		}
		target := subject.Targets[0]
		if subject.ClauseID > 0 && target.Exact && target.Cardinality.Min == 1 && target.Cardinality.Max == 1 &&
			enteredNounCompatible(reference.SubjectNoun, target.Selection.Kind) {
			reference.DeclaredPermanent.ClauseID = subject.ClauseID
		}
	}
	for _, effect := range effects {
		for _, list := range [][]Reference{effect.References, effect.SubjectReferences} {
			for i := range list {
				for _, canonical := range references {
					if list[i].NodeID == canonical.NodeID {
						list[i].DeclaredPermanent = canonical.DeclaredPermanent
						break
					}
				}
			}
		}
	}
}

func declaredPermanentProjection(reference Reference, effects []*EffectSyntax) bool {
	for _, effect := range effects {
		if (effect.Context == EffectContextReferencedObjectOwner ||
			effect.Context == EffectContextReferencedObjectController) &&
			len(effect.References) > 0 && effect.References[0].NodeID == reference.NodeID {
			return true
		}
	}
	return false
}

func declaredPermanentAction(effect *EffectSyntax) bool {
	switch effect.Kind {
	case EffectReturn:
		return effect.ToZone == zone.Hand
	case EffectDestroy, EffectTap, EffectUntap, EffectModifyPT, EffectGainControl, EffectGrantKeyword:
		return true
	default:
		return false
	}
}

func declaredPermanentReference(reference Reference) bool {
	if reference.SubjectNoun == ObjectNounCard || reference.SubjectNoun == ObjectNounPlayer ||
		reference.SubjectNoun == ObjectNounSpell {
		return false
	}
	return reference.Kind == ReferenceThatObject && reference.Pronoun == PronounUnknown ||
		reference.Kind == ReferencePronoun && (reference.Pronoun == PronounIt || reference.Pronoun == PronounIts)
}

func declaredPermanentSelection(selection SelectionSyntax) bool {
	if selection.Zone != zone.None && selection.Zone != zone.Battlefield {
		return false
	}
	switch selection.Kind {
	case SelectionPermanent, SelectionCreature, SelectionArtifact, SelectionEnchantment,
		SelectionLand, SelectionPlaneswalker, SelectionBattle:
		return true
	default:
		return false
	}
}
