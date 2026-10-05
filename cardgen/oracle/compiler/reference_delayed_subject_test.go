package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestDelayedSubjectProjectionUsesTypedIdentity(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		kind    parser.DelayedSubjectKind
		binding ReferenceBinding
	}{
		{parser.DelayedSubjectSource, ReferenceBindingSource},
		{parser.DelayedSubjectTarget, ReferenceBindingTarget},
		{parser.DelayedSubjectProduct, ReferenceBindingPriorInstructionResult},
		{parser.DelayedSubjectEvent, ReferenceBindingEventPermanent},
		{parser.DelayedSubjectEventRelated, ReferenceBindingEventRelatedPermanent},
		{parser.DelayedSubjectUnsupported, ReferenceBindingUnsupported},
	} {
		effects := []CompiledEffect{
			{ClauseID: 91, Text: "unrelated producer"},
			{ClauseID: 17, Text: "erased"},
			{ClauseID: 44, DelayedSubject: parser.DelayedSubjectOwnership{
				Kind: test.kind, ProducerClauseID: 17, TargetOccurrence: 3, ReferenceNodeIDs: []int{25},
			}},
		}
		reference := CompiledReference{NodeID: 25, Text: "misleading wording"}
		if !bindDelayedSubjectReference(&reference, effects) || reference.Binding != test.binding {
			t.Fatalf("reference = %+v, want binding %v", reference, test.binding)
		}
		if test.kind == parser.DelayedSubjectProduct && reference.PriorInstruction != 1 ||
			test.kind == parser.DelayedSubjectTarget && reference.Occurrence != 3 {
			t.Fatal("projection guessed adjacency or target zero")
		}
		spell := CompiledReference{NodeID: 26, Binding: ReferenceBindingEventStackObject}
		if bindDelayedSubjectReference(&spell, effects) || spell.Binding != ReferenceBindingEventStackObject {
			t.Fatal("delayed subject overrode an unrelated authoritative spell binding")
		}
	}
}
