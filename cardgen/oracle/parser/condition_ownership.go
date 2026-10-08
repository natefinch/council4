package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// ConditionScope distinguishes a clause from a single printed check governing
// several clauses. Unmodeled grammar retains an explicit unsupported scope.
type ConditionScope uint8

// Condition scopes retain absent or ambiguous ownership rather than guessing.
const (
	ConditionScopeUnknown ConditionScope = iota
	ConditionScopeClause
	ConditionScopeGroup
	ConditionScopeUnsupported
)

// ConditionOwnership names grammatical owners, not instruction positions.
type ConditionOwnership struct {
	DelayedBody      bool           `json:",omitempty"`
	Scope            ConditionScope `json:",omitempty"`
	ClauseIDs        []int          `json:",omitempty"`
	ReferenceNodeIDs []int          `json:",omitempty"`
	// ResultProducerClauseID names an earlier resolving action, never a cost.
	ResultProducerClauseID int `json:",omitempty"`
	// ResultSubjectReferenceNodeID identifies the specific countered spell.
	ResultSubjectReferenceNodeID  int `json:",omitempty"`
	ResultSubjectTargetOccurrence int `json:",omitempty"`
}

func emitAbilityConditionOwnership(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		emitConditionOwnership(ability.Sentences, ability.ConditionSegments, ability.SemanticReferences)
		emitCounterResultOwnership(ability.Sentences, ability.ConditionSegments, ability.ConditionClauses, ability.SemanticReferences)
		emitResultConditionOwnership(ability.Sentences, ability.ConditionSegments, ability.ConditionClauses)
		if ability.Modal != nil {
			for j := range ability.Modal.Options {
				mode := &ability.Modal.Options[j]
				emitConditionOwnership(mode.Sentences, mode.ConditionSegments, mode.SemanticReferences)
				emitCounterResultOwnership(mode.Sentences, mode.ConditionSegments, mode.ConditionClauses, mode.SemanticReferences)
				emitResultConditionOwnership(mode.Sentences, mode.ConditionSegments, mode.ConditionClauses)
			}
		}
	}
}

func emitConditionOwnership(sentences []Sentence, segments []ConditionSegment, references []Reference) {
	effects := conditionEffects(sentences)
	for ei, effect := range effects {
		effect.ClauseID = ei + 1
	}
	for ci := range segments {
		segment := &segments[ci]
		for _, reference := range references {
			if parserSpanContains(segment.Span, reference.Span) {
				segment.Ownership.ReferenceNodeIDs = append(segment.Ownership.ReferenceNodeIDs, reference.NodeID)
			}
		}
		var owners []*EffectSyntax
		for _, effect := range effects {
			if parserSpanContains(effect.Span, segment.Span) {
				owners = append(owners, effect)
			}
		}
		switch len(owners) {
		case 0:
			segment.Ownership.Scope = ConditionScopeUnknown
		case 1:
			segment.Ownership.Scope = ConditionScopeClause
			segment.Ownership.ClauseIDs = []int{owners[0].ClauseID}
			segment.Ownership.DelayedBody = delayedConditionEvaluation(owners[0], segment, sentences)
			if owners[0].Kind == EffectCantBeBlocked && owners[0].VerbSpan != (shared.Span{}) &&
				segment.Span.Start.Offset > owners[0].VerbSpan.Start.Offset {
				// A trailing restriction predicate is not proved to be a resolving grant.
				segment.Ownership.Scope = ConditionScopeUnsupported
			}
		default:
			segment.Ownership.Scope = ConditionScopeGroup
			for _, effect := range owners {
				if effect.Span != owners[0].Span || effect.VerbSpan.Start.Offset == 0 ||
					segment.Span.End.Offset > effect.VerbSpan.Start.Offset {
					segment.Ownership.Scope = ConditionScopeUnsupported
					segment.Ownership.ClauseIDs = nil
					break
				}
				segment.Ownership.ClauseIDs = append(segment.Ownership.ClauseIDs, effect.ClauseID)
			}
		}
	}
}

func conditionEffects(sentences []Sentence) []*EffectSyntax {
	var effects []*EffectSyntax
	var appendEffects func([]EffectSyntax)
	appendEffects = func(syntaxes []EffectSyntax) {
		for ei := range syntaxes {
			effect := &syntaxes[ei]
			if len(effect.RepeatBody) > 0 {
				appendEffects(effect.RepeatBody)
			} else {
				effects = append(effects, effect)
			}
		}
	}
	for si := range sentences {
		appendEffects(sentences[si].Effects)
	}
	return effects
}

func parserSpanContains(outer, inner shared.Span) bool {
	return inner.Start.Offset >= outer.Start.Offset && inner.End.Offset <= outer.End.Offset
}
