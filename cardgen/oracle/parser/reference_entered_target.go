package parser

import "github.com/natefinch/council4/mtg/game/zone"

// The action's own predicate names its input; a later subject names its output.
func bindEnteredTargetReferences(effects []*EffectSyntax, references []Reference, conditions []ConditionSegment) {
	for i := range references {
		reference := &references[i]
		if reference.ProducerClauseID > 0 || !libraryCardReference(*reference) ||
			reference.SubjectNoun == ObjectNounSpell || reference.SubjectNoun == ObjectNounPlayer {
			continue
		}
		var latest *EffectSyntax
		for _, effect := range effects {
			if effect.VerbSpan.End.Offset >= reference.Span.Start.Offset ||
				parserSpanContains(effect.ClauseSpan, reference.Span) ||
				libraryReferenceIsOwnActionCondition(*reference, effect.ClauseID, conditions) {
				continue
			}
			if len(effect.Targets) > 0 || enteredSubjectBarrier(effect) {
				latest = effect
			}
		}
		if latest == nil || !latest.Exact || latest.Negated || latest.DelayedTiming != DelayedTimingNone ||
			(latest.Kind != EffectPut && latest.Kind != EffectReturn) || latest.ToZone != zone.Battlefield ||
			len(latest.Targets) != 1 {
			continue
		}
		target := latest.Targets[0]
		if target.Selection.Kind == SelectionPlayer || target.Selection.Kind == SelectionOpponent ||
			target.Selection.Zone == zone.None ||
			target.Selection.Zone == zone.Battlefield || target.Cardinality.Min != 1 || target.Cardinality.Max != 1 ||
			!enteredNounCompatible(reference.SubjectNoun, target.Selection.Kind) ||
			closerSourceAntecedent(*reference, latest, references) {
			continue
		}
		reference.ProducerClauseID = latest.ClauseID
	}
}

// A typed noun naming another kind of object ("that land" after a returned
// creature card) is not the entered product.
func enteredNounCompatible(noun ObjectNoun, target SelectionKind) bool {
	if noun == ObjectNounUnknown || noun == ObjectNounCard || noun == ObjectNounPermanent {
		return true
	}
	if target == SelectionUnknown || target == SelectionCard || target == SelectionPermanent {
		return true
	}
	return selectionKindForNoun(noun) == target
}

// A bare pronoun refers to an explicit source antecedent named after the
// producing action ("This creature gets +1/+1 ... If it was an Elf").
func closerSourceAntecedent(reference Reference, producer *EffectSyntax, references []Reference) bool {
	if reference.Kind != ReferencePronoun {
		return false
	}
	for _, candidate := range references {
		if (candidate.Kind == ReferenceThisObject || candidate.Kind == ReferenceSelfName) &&
			candidate.Span.Start.Offset > producer.VerbSpan.End.Offset &&
			candidate.Span.Start.Offset < reference.Span.Start.Offset {
			return true
		}
	}
	return false
}

func enteredSubjectBarrier(effect *EffectSyntax) bool {
	switch effect.Kind {
	case EffectPut, EffectReturn:
		return effect.ToZone != zone.None
	case EffectExile, EffectReveal, EffectLookAtLibraryTop,
		EffectDig, EffectSearch, EffectChoosePermanent, EffectCreate,
		EffectManifestDread, EffectSacrifice, EffectDestroy:
		return true
	default:
		return false
	}
}
