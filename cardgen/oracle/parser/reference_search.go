package parser

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

// A completed singular search owns the arrived permanent across independent
// clauses. A new object subject closes that grammatical frame.
func bindSearchedPermanentReferences(effects []*EffectSyntax, references []Reference, conditions []ConditionSegment) {
	for ri := range references {
		reference := &references[ri]
		if reference.ProducerClauseID != 0 || reference.CardIdentity || !libraryCardReference(*reference) {
			continue
		}
		var producer *EffectSyntax
		groupEnd := -1
		for ei, effect := range effects {
			if effect.VerbSpan.Start.Offset >= reference.Span.Start.Offset ||
				parserSpanContains(effect.ClauseSpan, reference.Span) ||
				libraryReferenceIsOwnActionCondition(*reference, effect.ClauseID, conditions) {
				continue
			}
			if effect.Kind == EffectSearch {
				producer, groupEnd = nil, -1
				if end, ok := searchedGroupEnd(effects, ei); ok &&
					effects[end-1].VerbSpan.End.Offset < reference.Span.Start.Offset {
					producer, groupEnd = effect, end
				}
				continue
			}
			if producer != nil && ei < groupEnd {
				continue
			}
			if len(effect.Targets) != 0 || enteredSubjectBarrier(effect) {
				producer, groupEnd = nil, -1
			}
		}
		if producer != nil && !closerSourceAntecedent(*reference, producer, references) {
			for _, effect := range effects {
				if slices.ContainsFunc(effect.References, func(r Reference) bool { return r.NodeID == reference.NodeID }) &&
					!searchConsumerPrefixSupported(effect, conditions) {
					effect.Exact = false
					effect.TapUntapReferenceObjectClean = false
					producer = nil
				}
			}
			if producer == nil {
				continue
			}
			reference.ProducerClauseID = producer.ClauseID
		}
	}
}

func searchedGroupEnd(effects []*EffectSyntax, start int) (int, bool) {
	search := effects[start]
	if search.Kind != EffectSearch || search.Negated || search.DelayedTiming != DelayedTimingNone ||
		search.Context != EffectContextController || len(search.Targets) != 0 {
		return 0, false
	}
	putIndex := start + 1
	if putIndex < len(effects) && effects[putIndex].Kind == EffectReveal {
		putIndex++
	}
	if putIndex+1 >= len(effects) {
		return 0, false
	}
	put, shuffle := effects[putIndex], effects[putIndex+1]
	if put.Kind != EffectPut || put.ToZone != zone.Battlefield && put.ToZone != zone.Hand ||
		put.Negated || put.DelayedTiming != DelayedTimingNone || len(put.Targets) != 0 ||
		len(put.References) != 1 ||
		(!libraryCardReference(put.References[0]) &&
			put.References[0].Pronoun != PronounThem && put.References[0].Pronoun != PronounThose) ||
		shuffle.Kind != EffectShuffle || shuffle.Connection != EffectConnectionThen {
		return 0, false
	}
	return putIndex + 2, true
}

func searchConsumerPrefixSupported(effect *EffectSyntax, conditions []ConditionSegment) bool {
	for _, token := range effect.Tokens {
		if token.Span.Start.Offset >= effect.VerbSpan.Start.Offset {
			break
		}
		if token.Kind == shared.Comma || parserSpanContains(effect.ConnectionSpan, token.Span) {
			continue
		}
		owned := false
		for _, condition := range conditions {
			owned = owned || slices.Contains(condition.Ownership.ClauseIDs, effect.ClauseID) &&
				parserSpanContains(condition.Span, token.Span)
		}
		for _, reference := range effect.SubjectReferences {
			owned = owned || parserSpanContains(reference.Span, token.Span)
		}
		if !owned {
			return false
		}
	}
	return true
}
