package parser

import (
	"reflect"
	"testing"
)

func TestParseConditionOwnership(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		text   string
		scopes []ConditionScope
		owners [][]int
	}{
		{"If you have no cards in hand, draw a card, then draw a card.", []ConditionScope{ConditionScopeGroup}, [][]int{{1, 2}}},
		{"Draw a card. Unless you control a Villain, you lose 2 life.", []ConditionScope{ConditionScopeClause}, [][]int{{2}}},
		{"If you have no cards in hand, draw a card. If you have no cards in hand, draw a card.", []ConditionScope{ConditionScopeClause, ConditionScopeClause}, [][]int{{1}, {2}}},
		{"If you have no cards in hand, draw a card, then draw a card. If you control a creature, draw a card, then discard a card.", []ConditionScope{ConditionScopeGroup, ConditionScopeGroup}, [][]int{{1, 2}, {3, 4}}},
		{"Draw a card, then discard a card unless you control a Villain.", []ConditionScope{ConditionScopeUnsupported}, [][]int{nil}},
	} {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(tt.text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 || len(document.Abilities) != 1 {
				t.Fatalf("parse diagnostics=%#v abilities=%d", diagnostics, len(document.Abilities))
			}
			conditions := document.Abilities[0].ConditionSegments
			if len(conditions) != len(tt.scopes) {
				t.Fatalf("conditions=%d, want %d", len(conditions), len(tt.scopes))
			}
			for i, condition := range conditions {
				if condition.Ownership.Scope != tt.scopes[i] ||
					!reflect.DeepEqual(condition.Ownership.ClauseIDs, tt.owners[i]) {
					t.Fatalf("condition[%d] ownership=%#v", i, condition.Ownership)
				}
			}
		})
	}
}

func TestParseConditionReferenceOwnership(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("Counter target spell unless its controller pays {3}. Draw a card.",
		Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	ability := document.Abilities[0]
	condition := ability.ConditionSegments[0]
	if len(condition.Ownership.ReferenceNodeIDs) != 1 {
		t.Fatalf("ownership=%#v, want the payment's subject reference", condition.Ownership)
	}
	id := condition.Ownership.ReferenceNodeIDs[0]
	for _, reference := range ability.SemanticReferences {
		if reference.NodeID == id && reference.Pronoun == PronounIts {
			return
		}
	}
	t.Fatal("payment ownership did not identify its actual subject reference")
}
