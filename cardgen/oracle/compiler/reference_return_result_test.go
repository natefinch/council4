package compiler

import "testing"

func TestReturnedPermanentConditionIdentity(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Return target creature card from your graveyard to the battlefield. If it's an Angel, put two +1/+1 counters on it.",
		"Exile target creature you control, then return that card to the battlefield under its owner's control. If it's a Spirit, put a +1/+1 counter on it.",
		"{2}: Exile this creature, then return it to the battlefield under its owner's control. If it's an Elf, put a +1/+1 counter on it.",
	} {
		compilation, _ := compileSource(text, pipelineContext{CardName: "Returned Subject"})
		content := compilation.Abilities[0].Content
		condition := content.Conditions[0]
		if condition.Predicate != ConditionPredicateObjectMatches || condition.ObjectBinding != ReferenceBindingPriorInstructionResult {
			t.Fatalf("condition=%+v, want prior-result object predicate", condition)
		}
		if condition.ObjectReference == nil || condition.ObjectReference.NodeID != condition.SubjectRefID {
			t.Fatal("parser subject identity was lost")
		}
	}
}

func TestReturnedPermanentConditionExplicitSubjectPrecedence(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, text string
		binding    ReferenceBinding
	}{
		{"event noun", "Whenever a land enters, return target creature card from your graveyard to the battlefield. If that land is an Island, draw a card.", ReferenceBindingEventPermanent},
		{"returned pronoun", "Whenever a land enters, return target creature card from your graveyard to the battlefield. If it's an Elf, draw a card.", ReferenceBindingPriorInstructionResult},
		{"later source", "{2}: Return target creature card from your graveyard to the battlefield. This creature gets +1/+1 until end of turn. If it was an Elf, draw a card.", ReferenceBindingSource},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			compilation, _ := compileSource(test.text, pipelineContext{CardName: "Subject Precedence"})
			condition := compilation.Abilities[0].Content.Conditions[0]
			if condition.Predicate != ConditionPredicateObjectMatches ||
				condition.ObjectBinding != test.binding || condition.ObjectReference == nil ||
				condition.ObjectReference.NodeID != condition.SubjectRefID {
				t.Fatalf("condition=%+v, want binding=%v and preserved subject identity", condition, test.binding)
			}
		})
	}
}
