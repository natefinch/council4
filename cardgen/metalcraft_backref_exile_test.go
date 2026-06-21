package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerDispatchMetalcraftBackReferenceExile proves the two-paragraph
// removal spell Dispatch lowers to a single spell ability whose targeted-creature
// tap is followed by a back-reference exile gated on the Metalcraft control-count
// condition. The trailing "exile that creature" paragraph carries no target of
// its own; the spell-face combiner folds it onto the leading tap paragraph so the
// exile addresses the same target index.
func TestLowerDispatchMetalcraftBackReferenceExile(t *testing.T) {
	t.Parallel()

	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Dispatch",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Tap target creature.\nMetalcraft — If you control three or more artifacts, exile that creature.",
	})

	if !face.SpellAbility.Exists {
		t.Fatal("Dispatch did not lower to a spell ability")
	}
	content := face.SpellAbility.Val
	if len(content.Modes) != 1 {
		t.Fatalf("modes = %d, want one", len(content.Modes))
	}
	mode := content.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %d, want one creature target", len(mode.Targets))
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want tap then conditional exile", mode.Sequence)
	}

	tap, ok := mode.Sequence[0].Primitive.(game.Tap)
	if !ok || tap.Object != game.TargetPermanentReference(0) {
		t.Fatalf("first instruction = %#v, want tap of target zero", mode.Sequence[0])
	}
	if mode.Sequence[0].Condition.Exists {
		t.Fatal("tap must be unconditional")
	}

	exileInstruction := mode.Sequence[1]
	exile, ok := exileInstruction.Primitive.(game.Exile)
	if !ok || exile.Object != game.TargetPermanentReference(0) {
		t.Fatalf("second instruction = %#v, want exile of target zero", exileInstruction)
	}
	if !exileInstruction.Condition.Exists {
		t.Fatal("exile must be gated by the Metalcraft condition")
	}
	gate := exileInstruction.Condition.Val.Condition
	if !gate.Exists || !gate.Val.ControlsMatching.Exists {
		t.Fatalf("exile gate = %#v, want a control-count condition", exileInstruction.Condition.Val)
	}
	controls := gate.Val.ControlsMatching.Val
	if controls.MinCount != 3 ||
		len(controls.Selection.RequiredTypes) != 1 ||
		controls.Selection.RequiredTypes[0] != types.Artifact {
		t.Fatalf("control-count = %#v, want three or more artifacts", controls)
	}
}
