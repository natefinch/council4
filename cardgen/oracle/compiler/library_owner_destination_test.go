package compiler

import "testing"

func TestLibraryOwnerDestinationHasDistinctTypedProvenance(t *testing.T) {
	text := "Look at the top card of target player's library. If it's a nonland card, you may pay 2 life. If you do, put it into that player's graveyard."
	compilation, diagnostics := compileSource(text, pipelineContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	content := compilation.Abilities[0].Content
	effect := content.Effects[2]
	if !effect.Exact || !effect.LibraryOwnerDestination || len(effect.References) != 2 {
		t.Fatalf("destination: exact=%v owner=%v source=%v refs=%#v",
			effect.Exact, effect.LibraryOwnerDestination, effect.CardSource, effect.References)
	}
	for _, reference := range effect.References {
		want := ReferenceBindingPriorInstructionResult
		if reference.Kind == ReferenceThatPlayer {
			want = ReferenceBindingLibraryOwner
		}
		if reference.Binding != want || reference.PriorInstruction != 0 ||
			reference.ProducerClauseID != content.Effects[0].ClauseID {
			t.Fatalf("reference=%#v, want binding=%v exact producer=%d",
				reference, want, content.Effects[0].ClauseID)
		}
	}
}
