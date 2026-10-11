package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
)

func TestSequenceProductIndicesFollowExactClauseIdentity(t *testing.T) {
	for _, test := range []struct {
		name    string
		binding compiler.ReferenceBinding
		id      int
		want    int
	}{
		{"clipped product", compiler.ReferenceBindingPriorInstructionResult, 40, 0},
		{"later product", compiler.ReferenceBindingPriorInstructionResult, 80, 1},
		{"legacy binding", compiler.ReferenceBindingPriorInstructionResult, 0, 7},
		{"authoritative event", compiler.ReferenceBindingEventPermanent, 40, 7},
		{"authoritative target", compiler.ReferenceBindingTarget, 40, 7},
		{"authoritative source", compiler.ReferenceBindingSource, 40, 7},
	} {
		t.Run(test.name, func(t *testing.T) {
			ref := compiler.CompiledReference{Binding: test.binding, ProducerClauseID: test.id, PriorInstruction: 7}
			original := compiler.AbilityContent{
				References: []compiler.CompiledReference{ref},
				Conditions: []compiler.CompiledCondition{{ObjectBinding: ref.Binding, ObjectReference: &ref}},
				Effects: []compiler.CompiledEffect{
					{ClauseID: 40},
					{ClauseID: 80, References: []compiler.CompiledReference{ref}, SubjectReferences: []compiler.CompiledReference{ref}},
				},
			}
			actual, ok := normalizeSequenceProductReferences(original)
			if !ok {
				t.Fatal("unique typed product was refused")
			}
			for _, reference := range []compiler.CompiledReference{
				actual.References[0], *actual.Conditions[0].ObjectReference,
				actual.Effects[1].References[0], actual.Effects[1].SubjectReferences[0],
			} {
				if reference.Binding != ref.Binding || reference.PriorInstruction != test.want || reference.ProducerClauseID != ref.ProducerClauseID {
					t.Fatalf("reference changed role or lost exact producer: %#v", reference)
				}
			}
			if original.References[0].PriorInstruction != 7 || original.Conditions[0].ObjectReference.PriorInstruction != 7 ||
				original.Effects[1].References[0].PriorInstruction != 7 || original.Effects[1].SubjectReferences[0].PriorInstruction != 7 {
				t.Fatal("normalization mutated the original enclosing content")
			}
		})
	}
}

func TestSequenceProductIndicesRejectMissingOrDuplicateProducer(t *testing.T) {
	for _, effects := range [][]compiler.CompiledEffect{
		{{ClauseID: 12}},
		{{ClauseID: 40}, {ClauseID: 40}},
	} {
		content := compiler.AbilityContent{
			Effects: effects,
			References: []compiler.CompiledReference{{
				Binding: compiler.ReferenceBindingPriorInstructionResult, ProducerClauseID: 40,
			}},
		}
		if _, ok := normalizeSequenceProductReferences(content); ok {
			t.Fatal("missing or duplicate ClauseID manufactured a product index")
		}
	}
}
