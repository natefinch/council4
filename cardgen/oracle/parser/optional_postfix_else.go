package parser

import "slices"

// OptionalPostfixElseOwnership distinguishes an optional action's postfix
// condition from a predicate-leading optional consequence.
type OptionalPostfixElseOwnership struct {
	ActionClauseID  int `json:",omitempty"`
	ConditionNodeID int `json:",omitempty"`
}

func emitOptionalPostfixElseOwnership(effects []*EffectSyntax, segments []ConditionSegment) {
	for ei, effect := range effects {
		if ei == 0 || effect.Connection != EffectConnectionOtherwise {
			continue
		}
		previous := effects[ei-1]
		for _, action := range effects[:ei] {
			if !action.Optional || len(action.OptionalActionClauseIDs) == 0 ||
				action.OptionalActionClauseIDs[len(action.OptionalActionClauseIDs)-1] != previous.ClauseID {
				continue
			}
			conditionID, count := -1, 0
			for _, segment := range segments {
				if segment.Kind != ConditionIntroIf || segment.Intervening ||
					segment.Span.Start.Offset <= action.VerbSpan.Start.Offset ||
					!slices.ContainsFunc(segment.Ownership.ClauseIDs, func(id int) bool {
						return slices.Contains(action.OptionalActionClauseIDs, id)
					}) {
					continue
				}
				count++
				conditionID = segment.NodeID
			}
			if count == 0 {
				continue
			}
			if count != 1 {
				conditionID = -1
			}
			for _, consequence := range effects[ei:] {
				if consequence.Span != effect.Span {
					break
				}
				consequence.OptionalPostfixElse = OptionalPostfixElseOwnership{
					ActionClauseID: action.ClauseID, ConditionNodeID: conditionID,
				}
			}
		}
	}
}
