package parser

import "testing"

// TestParseRapaciousGuestLeaveLifeLoss verifies the Rapacious Guest departure
// clause "target opponent loses life equal to its power" parses to a targeted
// life-loss effect whose amount is the source-power dynamic ("its power"). The
// parser stays text-blind: it records that the amount reads a power
// characteristic and pins the "its" referent span, leaving the referent binding
// to the compiler.
func TestParseRapaciousGuestLeaveLifeLoss(t *testing.T) {
	t.Parallel()
	source := "When this creature leaves the battlefield, target opponent loses life equal to its power."
	document, diagnostics := Parse(source, Context{CardName: "Rapacious Guest"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	var lose *EffectSyntax
	for ai := range document.Abilities {
		for si := range document.Abilities[ai].Sentences {
			sentence := &document.Abilities[ai].Sentences[si]
			for ei := range sentence.Effects {
				if sentence.Effects[ei].Kind == EffectLose {
					lose = &sentence.Effects[ei]
				}
			}
		}
	}
	if lose == nil {
		t.Fatalf("no lose-life effect parsed from %q", source)
	}
	if lose.Context != EffectContextTarget {
		t.Fatalf("lose context = %v, want EffectContextTarget", lose.Context)
	}
	if lose.Amount.DynamicKind != EffectDynamicAmountSourcePower {
		t.Fatalf("lose amount dynamic kind = %v, want EffectDynamicAmountSourcePower", lose.Amount.DynamicKind)
	}
	if !lose.Exact {
		t.Fatal("lose Exact = false, want true")
	}
	if lose.Amount.ReferenceSpan.End.Offset <= lose.Amount.ReferenceSpan.Start.Offset {
		t.Fatal("lose amount reference span is empty, want the \"its\" span")
	}
	if lose.Amount.ReferenceNodeID < 0 {
		t.Fatalf("lose amount reference node ID = %d, want the \"its\" reference NodeID", lose.Amount.ReferenceNodeID)
	}
}
