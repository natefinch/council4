package parser

import "slices"

func resultActorProducer(effects []*EffectSyntax, sentences []Sentence, references []Reference, segment *ConditionSegment, first int) (int, bool) {
	var actor *Reference
	for ri := range references {
		reference := &references[ri]
		if !slices.Contains(segment.Ownership.ReferenceNodeIDs, reference.NodeID) ||
			(reference.Kind != ReferenceThatPlayer && (reference.Kind != ReferencePronoun || reference.Pronoun != PronounThey)) {
			continue
		}
		if actor != nil {
			return -1, true
		}
		actor = reference
	}
	if actor == nil {
		return -1, false
	}
	ids := make(map[int]bool)
	for _, effect := range effects {
		if effect.ClauseID <= 0 || ids[effect.ClauseID] {
			return -1, true
		}
		ids[effect.ClauseID] = true
	}
	segment.Ownership.ResultSubjectReferenceNodeID = actor.NodeID
	segment.Ownership.ResultSubjectTargetOccurrence = -1
	var targets []TargetSyntax
	for _, sentence := range sentences {
		targets = append(targets, sentence.Targets...)
	}
	occurrence, unique := resultActorTarget(targets, *actor)
	if !unique && len(targets) != 0 {
		return -1, true
	}
	segment.Ownership.ResultSubjectTargetOccurrence = occurrence
	for ei := first - 1; ei >= 0; ei-- {
		effect := effects[ei]
		if occurrence >= 0 && slices.ContainsFunc(effect.Targets, func(target TargetSyntax) bool {
			return target.Order == targets[occurrence].Order
		}) {
			switch targets[occurrence].Selection.Kind {
			case SelectionSpell:
				if effect.Kind != EffectCounter || effect.Payment.Form != EffectPaymentFormUnless ||
					effect.Payment.Payer != EffectPaymentPayerTargetController {
					return -1, true
				}
			case SelectionPlayer, SelectionOpponent:
				if effect.Context != EffectContextTarget {
					return -1, true
				}
			default:
				return -1, true
			}
			return ei, true
		}
		if effect.Context == EffectContextController {
			continue
		}
		if effect.Context == EffectContextReferencedPlayer || effect.Context == EffectContextEventPlayer ||
			effect.Context == EffectContextReferencedObjectController {
			subjects := slices.DeleteFunc(slices.Clone(effect.SubjectReferences), func(reference Reference) bool {
				return !resultPlayerActor(reference) &&
					(effect.Context != EffectContextReferencedObjectController ||
						reference.Kind != ReferencePronoun || reference.Pronoun != PronounIts)
			})
			if len(subjects) != 1 {
				return -1, true
			}
			var canonical *Reference
			for ri := range references {
				reference := &references[ri]
				if reference.NodeID != subjects[0].NodeID {
					continue
				}
				if canonical != nil || reference.Kind != subjects[0].Kind || reference.Pronoun != subjects[0].Pronoun {
					return -1, true
				}
				canonical = reference
			}
			if canonical == nil || !effect.Order.Contains(canonical.Order) {
				return -1, true
			}
			subjectTarget, unique := resultActorTarget(targets, *canonical)
			if !unique && len(targets) != 0 {
				return -1, true
			}
			if subjectTarget == occurrence && canonical.Order.End <= actor.Order.Start {
				return ei, true
			}
			continue
		}
		if effect.Context == EffectContextPriorSubject && len(effect.SubjectReferences) == 0 {
			return -1, true
		}
	}
	return -1, true
}

func resultPlayerActor(reference Reference) bool {
	return reference.Kind == ReferenceThatPlayer ||
		(reference.Kind == ReferencePronoun && reference.Pronoun == PronounThey)
}

func resultActorTarget(targets []TargetSyntax, actor Reference) (int, bool) {
	occurrence := -1
	ambiguous := false
	for ti, target := range targets {
		if target.Order.End > actor.Order.Start {
			continue
		}
		if occurrence < 0 || target.Order.Start > targets[occurrence].Order.Start {
			occurrence = ti
			ambiguous = false
		} else if target.Order.Start == targets[occurrence].Order.Start {
			ambiguous = true
		}
	}
	if occurrence < 0 || ambiguous {
		return -1, false
	}
	target := targets[occurrence]
	if target.Cardinality.Min != 1 || target.Cardinality.Max != 1 {
		return -1, false
	}
	return occurrence, true
}
