package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceWellOfLostDreams covers Well of Lost Dreams:
// "Whenever you gain life, you may pay {X}, where X is less than or equal to the
// amount of life you gained. If you do, draw X cards." The bounded optional
// payment lowers to a controller PayRepeatedly that offers {1} up to the amount
// of life gained times — the MaxCount bound reads the triggering life-change
// quantity — and publishes the count, followed by a Draw of that many cards gated
// on the payment having succeeded.
func TestGenerateExecutableCardSourceWellOfLostDreams(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Well of Lost Dreams",
		Layout:     "normal",
		ManaCost:   "{4}",
		TypeLine:   "Artifact",
		OracleText: "Whenever you gain life, you may pay {X}, where X is less than or equal to the amount of life you gained. If you do, draw X cards.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "w")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Event:  game.EventLifeGained,",
		"Primitive: game.PayRepeatedly{",
		"Payer:  opt.Val(game.ControllerReference()),",
		"cost.O(1),",
		`PublishCount: "variable-pay-scaled-draw-count",`,
		"MaxCount: opt.Val(&game.DynamicAmount{",
		"Kind:       game.DynamicAmountEventLifeChange,",
		`PublishResult: game.ResultKey("variable-pay-scaled-draw-count"),`,
		"Primitive: game.Draw{",
		"Kind:      game.DynamicAmountChosenNumber,",
		`ResultKey: game.ResultKey("variable-pay-scaled-draw-count"),`,
		"Player: game.ControllerReference(),",
		"ResultGate: opt.Val(game.InstructionResultGate{",
		"Succeeded: game.TriTrue,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
