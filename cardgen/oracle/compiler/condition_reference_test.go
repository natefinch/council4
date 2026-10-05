package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestContextualObjectConditionSubject(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		text       string
		binding    ReferenceBinding
		occurrence int
		selector   SelectorKind
	}{
		{"adjacent card target", "When this creature enters, exile target card from a graveyard. If it was a creature card, draw a card.", ReferenceBindingTarget, 0, SelectorCard},
		{"nonadjacent card target", "When this creature enters, exile target card from a graveyard. You gain 1 life. If it was a creature card, draw a card.", ReferenceBindingTarget, 0, SelectorCard},
		{"contracted card target", "When this creature enters, exile target card from a graveyard. If it's a creature card, draw a card.", ReferenceBindingTarget, 0, SelectorCard},
		{"compound contracted card target", "When this creature enters, exile target card from a graveyard. If it's an artifact creature card, draw a card.", ReferenceBindingTarget, 0, SelectorCard},
		{"unsupported qualified contraction", "When this creature enters, exile target card from a graveyard. If it's a blue creature card, draw a card.", ReferenceBindingUnsupported, 0, SelectorUnknown},
		{"second target occurrence", "When this creature enters, tap target creature. Exile target card from a graveyard. If it was a land card, draw a card.", ReferenceBindingTarget, 1, SelectorCard},
		{"permanent target", "When this creature enters, destroy target creature. If that creature was a Human, draw a card.", ReferenceBindingTarget, 0, SelectorCreature},
		{"plural target is ambiguous", "When this creature enters, exile two target cards from a graveyard. If it was a creature card, draw a card.", ReferenceBindingUnsupported, 0, SelectorUnknown},
		{"omitted optional preceding target", "When this creature enters, put a +1/+1 counter on up to one target creature. Exile target card from a graveyard. If it was a creature card, draw a card.", ReferenceBindingUnsupported, 0, SelectorUnknown},
		{"two targets in one clause", "When this creature enters, exile target card from a graveyard and target card from another graveyard. If it was a creature card, draw a card.", ReferenceBindingUnsupported, 0, SelectorUnknown},
		{"event with no target", "Whenever another creature enters, if that creature was a Human, draw a card.", ReferenceBindingEventPermanent, 0, SelectorUnknown},
		{"event condition before target", "Whenever another creature enters, if that creature was a Human, tap target creature.", ReferenceBindingEventPermanent, 0, SelectorUnknown},
		{"event land is not creature target", "Whenever a land you control enters, tap target creature. If that land is an Island, draw a card.", ReferenceBindingEventPermanent, 0, SelectorUnknown},
		{"target land is not creature event", "Whenever another creature enters, tap target land. If that land is an Island, draw a card.", ReferenceBindingTarget, 0, SelectorLand},
		{"mismatched noun without proven event", "When this creature enters, tap target creature. If that land is an Island, draw a card.", ReferenceBindingUnsupported, 0, SelectorUnknown},
		{"competing land noun antecedents", "Whenever a land you control enters, tap target land. Tap target creature. If that land is an Island, draw a card.", ReferenceBindingUnsupported, 0, SelectorUnknown},
		{"source supersedes earlier target", "When this creature enters, tap target creature. This creature gets +1/+1 until end of turn. If it was a Human, draw a card.", ReferenceBindingSource, 0, SelectorUnknown},
		{"target supersedes earlier source", "When this creature enters, sacrifice this creature. Exile target card from a graveyard. If it was a creature card, draw a card.", ReferenceBindingTarget, 0, SelectorCard},
		{"batched event has no singular subject", "Whenever one or more creatures enter, if that creature was a Human, draw a card.", ReferenceBindingUnsupported, 0, SelectorUnknown},
		{"unbound subject", "If it was a creature card, draw a card.", ReferenceBindingUnsupported, 0, SelectorUnknown},
		{"resolution choice stays deferred", "When this creature enters, reveal the top card of your library. If it was a creature card, draw a card.", ReferenceBindingUnsupported, 0, SelectorUnknown},
		{"resolution choice supersedes earlier target", "When this creature enters, tap target creature. Reveal the top card of your library. If it was a creature card, draw a card.", ReferenceBindingUnsupported, 0, SelectorUnknown},
		{"resolution choice supersedes earlier source", "When this creature enters, sacrifice this creature. Reveal the top card of your library. If it was a creature card, draw a card.", ReferenceBindingUnsupported, 0, SelectorUnknown},
		{"nonadjacent choice supersedes earlier source", "When this creature enters, sacrifice this creature. Reveal the top card of your library. You gain 1 life. If it was a creature card, draw a card.", ReferenceBindingUnsupported, 0, SelectorUnknown},
		{"choice supersedes newer target and older source", "When this creature enters, sacrifice this creature. Exile target card from a graveyard. Reveal the top card of your library. You gain 1 life. If it was a creature card, draw a card.", ReferenceBindingUnsupported, 0, SelectorUnknown},
		{"blink incarnation is published", "When this creature enters, exile target creature. Return it to the battlefield under its owner's control. If that creature was a Human, draw a card.", ReferenceBindingPriorInstructionResult, 0, SelectorUnknown},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, _ := compileSource(test.text, pipelineContext{CardName: "Subject Test"})
			condition := compilation.Abilities[0].Content.Conditions[0]
			if test.binding == ReferenceBindingUnsupported {
				if condition.Predicate != ConditionPredicateUnsupported {
					t.Fatalf("condition = %#v, want unsupported", condition)
				}
				return
			}
			if condition.Predicate != ConditionPredicateObjectMatches ||
				condition.ObjectReference == nil ||
				condition.ObjectBinding != test.binding ||
				condition.ObjectReference.Binding != test.binding {
				t.Fatalf("condition = %#v, want full subject binding %v", condition, test.binding)
			}
			if condition.ObjectReference.NodeID != condition.SubjectRefID {
				t.Fatal("condition lost parser-owned subject identity")
			}
			if test.binding == ReferenceBindingTarget &&
				(condition.ObjectReference.Occurrence != test.occurrence ||
					condition.ObjectTarget == nil || condition.ObjectTarget.Selector.Kind != test.selector) {
				t.Fatalf("reference/target = %#v / %#v, want occurrence %d domain %v",
					condition.ObjectReference, condition.ObjectTarget, test.occurrence, test.selector)
			}
		})
	}
}

func TestContextualObjectConditionUsesTypedIdentity(t *testing.T) {
	t.Parallel()
	condition := CompiledCondition{ClauseIndex: 0, EventHistoryIndex: -1}
	recognizeCondition(&condition, []parser.ConditionClause{{
		Predicate:      parser.ConditionPredicateObjectMatches,
		ObjectBinding:  parser.ConditionObjectBindingEventPermanent,
		HasSubjectSpan: true,
		SubjectRefID:   17,
		Selection: parser.ConditionSelection{
			RequiredTypes: []parser.TriggerCardType{parser.TriggerCardTypeCreature},
		},
	}}, nil)
	references := []CompiledReference{
		{NodeID: 3, Binding: ReferenceBindingEventPermanent, Text: "it"},
		{NodeID: 17, Binding: ReferenceBindingTarget, Occurrence: 1, Text: "not Oracle", Order: shared.SourceOrder{Start: 20, End: 21}},
	}
	targets := []CompiledTarget{
		{Cardinality: TargetCardinality{Min: 1, Max: 1}, Selector: CompiledSelector{Kind: SelectorCreature}},
		{Cardinality: TargetCardinality{Min: 1, Max: 1}, Selector: CompiledSelector{Kind: SelectorCard}},
	}
	if !bindContextualObjectCondition(&condition, references, targets, nil, nil) ||
		condition.ObjectReference.NodeID != 17 || condition.ObjectReference.Occurrence != 1 {
		t.Fatalf("typed subject was not retained: %#v", condition)
	}
}
