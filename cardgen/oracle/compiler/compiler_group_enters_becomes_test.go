package compiler

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

// TestCompileGroupEntersBecomesReplacement proves Displaced Dinosaurs compiles to
// a replacement ability whose single effect carries the enters-becomes group
// modification with the historic filter, you-control scope, added Creature type
// and Dinosaur subtype, and 7/7 base P/T, and with no residual references.
func TestCompileGroupEntersBecomesReplacement(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"As a historic permanent you control enters, it becomes a 7/7 Dinosaur creature in addition to its other types.",
		pipelineContext{CardName: "Displaced Dinosaurs"},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(compilation.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(compilation.Abilities))
	}
	content := compilation.Abilities[0].Content
	if len(content.References) != 0 {
		t.Errorf("references = %#v, want none stripped for the becomes clause", content.References)
	}
	if len(content.Effects) != 1 {
		t.Fatalf("effects = %d, want 1", len(content.Effects))
	}
	if !content.Effects[0].EntersBecomesGroup() {
		t.Fatalf("effect is not an enters-becomes group: %#v", content.Effects[0])
	}
	mod := content.Effects[0].GroupEntryModification
	if !mod.Historic {
		t.Error("modification is not historic")
	}
	if len(mod.AddTypes) != 1 || mod.AddTypes[0] != types.Creature {
		t.Errorf("add types = %v, want [Creature]", mod.AddTypes)
	}
	if len(mod.AddSubtypes) != 1 || mod.AddSubtypes[0] != types.Dinosaur {
		t.Errorf("add subtypes = %v, want [Dinosaur]", mod.AddSubtypes)
	}
	if !mod.BasePower.Exists || mod.BasePower.Val != 7 ||
		!mod.BaseToughness.Exists || mod.BaseToughness.Val != 7 {
		t.Errorf("base P/T = %v/%v, want 7/7", mod.BasePower, mod.BaseToughness)
	}
}
