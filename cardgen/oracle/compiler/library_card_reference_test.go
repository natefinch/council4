package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompilerPreservesLibraryCardProducerOwnership(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Reveal the top card of your library. Draw a card. If it's a land card, put it onto the battlefield. Reveal the top card of your library. Scry 1. If it's a creature card, put the revealed card into your hand.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	content := compilation.Abilities[0].Content
	if len(content.Effects) != 6 || len(content.References) != 4 {
		t.Fatalf("effects=%d references=%#v", len(content.Effects), content.References)
	}
	for i, reference := range content.References {
		producer := 0
		if i > 1 {
			producer = 3
		}
		if reference.ProducerClauseID != content.Effects[producer].ClauseID {
			t.Errorf("reference %d lost its parser-owned producer: %#v", reference.NodeID, reference)
		}
		if reference.Binding != ReferenceBindingPriorInstructionResult || reference.PriorInstruction != producer {
			t.Errorf("reference %d did not bind the exact observed product: %#v", reference.NodeID, reference)
		}
	}
}

func TestLibraryCardBindingPreservesAuthoritativeSubjects(t *testing.T) {
	for _, binding := range []ReferenceBinding{
		ReferenceBindingSource, ReferenceBindingTarget, ReferenceBindingEventPermanent,
		ReferenceBindingEventCard, ReferenceBindingEventStackObject, ReferenceBindingEventPlayer,
		ReferenceBindingAmbiguous,
	} {
		reference := CompiledReference{
			NodeID: 17, ProducerClauseID: 3, Binding: binding, PriorInstruction: 23, Occurrence: 2,
		}
		effects := []CompiledEffect{{
			ClauseID: 3, Kind: EffectReveal, CardSource: parser.EffectCardSourceTopOfPlayerLibrary,
			Context: parser.EffectContextController, Exact: true,
			Amount: CompiledAmount{Known: true, Value: 1},
		}}
		if bindLibraryCardReference(&reference, effects) || reference.Binding != binding ||
			reference.NodeID != 17 || reference.Occurrence != 2 || reference.PriorInstruction != 23 {
			t.Fatalf("authoritative subject %v was replaced: %#v", binding, reference)
		}
	}
}
