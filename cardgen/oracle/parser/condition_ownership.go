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
	Scope            ConditionScope `json:",omitempty"`
	ClauseIDs        []int          `json:",omitempty"`
	ReferenceNodeIDs []int          `json:",omitempty"`
}

func emitAbilityConditionOwnership(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		emitConditionOwnership(ability.Sentences, ability.ConditionSegments, ability.SemanticReferences)
		if ability.Modal != nil {
			for j := range ability.Modal.Options {
				mode := &ability.Modal.Options[j]
				emitConditionOwnership(mode.Sentences, mode.ConditionSegments, mode.SemanticReferences)
			}
		}
	}
}

func emitConditionOwnership(sentences []Sentence, segments []ConditionSegment, references []Reference) {
	var effects []*EffectSyntax
	for si := range sentences {
		for ei := range sentences[si].Effects {
			effect := &sentences[si].Effects[ei]
			effect.ClauseID = len(effects) + 1
			effects = append(effects, effect)
		}
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

func parserSpanContains(outer, inner shared.Span) bool {
	return inner.Start.Offset >= outer.Start.Offset && inner.End.Offset <= outer.End.Offset
}
