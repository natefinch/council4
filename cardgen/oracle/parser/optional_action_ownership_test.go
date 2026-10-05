package parser

import (
	"reflect"
	"testing"
)

func TestOptionalActionScopeWithConditionsAndRiders(t *testing.T) {
	for _, tt := range []struct {
		name, text string
		groups     map[int][]int
	}{
		{"bare", "You may draw a card and gain 2 life. Scry 1.", map[int][]int{1: {1, 2}}},
		{"group condition", "If you have no cards in hand, you may draw a card and gain 2 life. Scry 1.", map[int][]int{1: {1, 2}}},
		{"new condition", "You may discard a card. If a land card was discarded this way, draw a card.", map[int][]int{1: {1}}},
		{"two decisions", "You may draw a card and gain 2 life, then you may draw a card and gain 3 life.", map[int][]int{1: {1, 2}, 3: {3, 4}}},
		{"independent rider", "You may untap target creature and gain control of it until end of turn. It gains haste until end of turn.", map[int][]int{1: {1, 2}}},
		{"aggregate outcome unresolved", "You may draw a card and gain 2 life. If you do, scry 1.", map[int][]int{1: {1, 2}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			document, diagnostics := Parse(tt.text, Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			ability := document.Abilities[0]
			for _, effect := range conditionEffects(ability.Sentences) {
				if !reflect.DeepEqual(effect.OptionalActionClauseIDs, tt.groups[effect.ClauseID]) {
					t.Fatalf("clause %d group=%v, want %v", effect.ClauseID, effect.OptionalActionClauseIDs, tt.groups[effect.ClauseID])
				}
			}
			if tt.name == "aggregate outcome unresolved" &&
				ability.ConditionSegments[0].Ownership.ResultProducerClauseID != 0 {
				t.Fatal("whole-group outcome was assigned to the last primitive")
			}
		})
	}
}
