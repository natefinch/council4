package game

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/opt"
)

func TestValidateDynamicResolutionPayment(t *testing.T) {
	t.Parallel()
	dynamic := DynamicAmount{
		Kind:       DynamicAmountObjectPower,
		Multiplier: 1,
		Object:     SourcePermanentReference(),
	}
	if err := ValidateInstructionSequence([]Instruction{{
		Primitive: Pay{Payment: ResolutionPayment{
			Payer:                  opt.Val(EventPlayerReference()),
			DynamicGenericManaCost: opt.Val(dynamic),
		}},
	}}); err != nil {
		t.Fatalf("dynamic payment validation error = %v", err)
	}
	if err := ValidateInstructionSequence([]Instruction{{
		Primitive: Pay{Payment: ResolutionPayment{
			ManaCost:               opt.Val(cost.Mana{cost.O(1)}),
			DynamicGenericManaCost: opt.Val(dynamic),
		}},
	}}); err == nil || !strings.Contains(err.Error(), "combine fixed and dynamic") {
		t.Fatalf("fixed plus dynamic payment error = %v", err)
	}
	if err := ValidateInstructionSequence([]Instruction{{
		Primitive: Pay{Payment: ResolutionPayment{}},
	}}); err == nil || !strings.Contains(err.Error(), "no cost") {
		t.Fatalf("empty payment error = %v", err)
	}
}
