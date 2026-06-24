package game

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
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
			DynamicGenericManaCost: opt.Val(&dynamic),
		}},
	}}); err != nil {
		t.Fatalf("dynamic payment validation error = %v", err)
	}
	if err := ValidateInstructionSequence([]Instruction{{
		Primitive: Pay{Payment: ResolutionPayment{
			ManaCost:               opt.Val(cost.Mana{cost.O(1)}),
			DynamicGenericManaCost: opt.Val(&dynamic),
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

func TestValidateMultipliedManaResolutionPayment(t *testing.T) {
	t.Parallel()
	multiplier := DynamicAmount{
		Kind:        DynamicAmountObjectCounters,
		Object:      SourcePermanentReference(),
		CounterKind: counter.Age,
	}
	valid := ResolutionPayment{
		ManaCost:           opt.Val(cost.Mana{cost.O(1), cost.U}),
		ManaCostMultiplier: opt.Val(&multiplier),
	}
	if err := ValidateInstructionSequence([]Instruction{{Primitive: Pay{Payment: valid}}}); err != nil {
		t.Fatalf("multiplied payment validation error = %v", err)
	}
	withoutBase := valid
	withoutBase.ManaCost = opt.V[cost.Mana]{}
	if err := ValidateInstructionSequence([]Instruction{{Primitive: Pay{Payment: withoutBase}}}); err == nil ||
		!strings.Contains(err.Error(), "requires a fixed mana cost") {
		t.Fatalf("multiplier without base error = %v", err)
	}
	withGenericDynamic := valid
	withGenericDynamic.DynamicGenericManaCost = opt.Val(&multiplier)
	if err := ValidateInstructionSequence([]Instruction{{Primitive: Pay{Payment: withGenericDynamic}}}); err == nil ||
		!strings.Contains(err.Error(), "cannot combine") {
		t.Fatalf("combined dynamic payment error = %v", err)
	}
}

func TestValidatePayRepeatedlyPublishesConsumableCount(t *testing.T) {
	t.Parallel()
	const countKey = ResultKey("paid-count")
	sequence := []Instruction{
		{Primitive: PayRepeatedly{
			Payment:      ResolutionPayment{ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G})},
			PublishCount: countKey,
		}},
		{Primitive: AddCounter{
			Amount:      Dynamic(DynamicAmount{Kind: DynamicAmountChosenNumber, ResultKey: countKey}),
			Object:      SourcePermanentReference(),
			CounterKind: counter.PlusOnePlusOne,
		}},
	}
	if err := ValidateInstructionSequence(sequence); err != nil {
		t.Fatalf("pay-repeatedly publish/consume validation error = %v", err)
	}
}

func TestValidatePayRepeatedlyRequiresPublishedCount(t *testing.T) {
	t.Parallel()
	if err := ValidateInstructionSequence([]Instruction{{Primitive: PayRepeatedly{
		Payment: ResolutionPayment{ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G})},
	}}}); err == nil || !strings.Contains(err.Error(), "published count key") {
		t.Fatalf("missing published count error = %v", err)
	}
}

func TestValidatePayRepeatedlyRequiresCost(t *testing.T) {
	t.Parallel()
	if err := ValidateInstructionSequence([]Instruction{{Primitive: PayRepeatedly{
		PublishCount: ResultKey("paid-count"),
	}}}); err == nil || !strings.Contains(err.Error(), "no cost") {
		t.Fatalf("missing cost error = %v", err)
	}
}
