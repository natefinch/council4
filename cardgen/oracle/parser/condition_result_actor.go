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
	segment.Ownership.ResultSubjectReferenceNodeID = actor.NodeID
	segment.Ownership.ResultSubjectTargetOccurrence = -1
	var targets []TargetSyntax
	for _, sentence := range sentences {
		targets = append(targets, sentence.Targets...)
	}
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
		return -1, true
	}
	target := targets[occurrence]
	if target.Cardinality.Min != 1 || target.Cardinality.Max != 1 {
		return -1, true
	}
	owner := -1
	for ei, effect := range effects {
		if !slices.ContainsFunc(effect.Targets, func(other TargetSyntax) bool { return other.Order == target.Order }) {
			continue
		}
		if owner >= 0 || ei >= first {
			return -1, true
		}
		owner = ei
	}
	if owner < 0 {
		return -1, true
	}
	for _, effect := range effects[owner+1 : first] {
		// A target declaration identifies the actor, not the outcome of a
		// later action by that actor. Do not reuse the declaration's receipt.
		if effect.Context == EffectContextReferencedObjectController ||
			(effect.Context == EffectContextPriorSubject && len(effect.SubjectReferences) == 0) ||
			slices.ContainsFunc(effect.SubjectReferences, func(reference Reference) bool {
				return reference.Kind == ReferenceThatPlayer ||
					(reference.Kind == ReferencePronoun && reference.Pronoun == PronounThey)
			}) {
			return -1, true
		}
	}
	effect := effects[owner]
	switch target.Selection.Kind {
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
	segment.Ownership.ResultSubjectTargetOccurrence = occurrence
	return owner, true
}
