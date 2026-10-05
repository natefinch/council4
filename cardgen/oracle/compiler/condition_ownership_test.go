package compiler

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompileConditionOwnershipFromTypedNodes(t *testing.T) {
	t.Parallel()
	ownership := parser.ConditionOwnership{
		Scope: parser.ConditionScopeGroup, ClauseIDs: []int{8, 21}, ReferenceNodeIDs: []int{19},
	}
	conditions := compileConditions([]parser.ConditionSegment{{
		Kind: parser.ConditionIntroIf, Ownership: ownership, NodeID: 7,
		ClauseIndex: -1, EventHistoryIndex: -1, Text: "not Oracle syntax",
	}}, nil, nil)
	if len(conditions) != 1 || !reflect.DeepEqual(conditions[0].Ownership, ownership) {
		t.Fatalf("conditions=%#v, want unchanged typed ownership", conditions)
	}
	effects := compileEffects([]parser.Sentence{{Effects: []parser.EffectSyntax{
		{ClauseID: 8, Kind: parser.EffectDraw, Optional: true, OptionalActionClauseIDs: []int{8, 21}},
		{ClauseID: 21, Kind: parser.EffectDiscard},
	}}})
	if len(effects) != 2 || effects[0].ClauseID != 8 || effects[1].ClauseID != 21 {
		t.Fatalf("effects=%#v, want parser identities unchanged", effects)
	}
	if !reflect.DeepEqual(effects[0].OptionalActionClauseIDs, []int{8, 21}) {
		t.Fatalf("optional group=%v, want unchanged parser identities", effects[0].OptionalActionClauseIDs)
	}
}
