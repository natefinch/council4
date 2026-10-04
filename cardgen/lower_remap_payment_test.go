package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/opt"
)

func TestCounterTaxRemapsPayerAndObject(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name: "Tax Rebase", Layout: "normal", TypeLine: "Instant",
		OracleText: "Target player gains 1 life. Counter target spell unless its controller pays {3}. You gain 2 life.",
	}
	assertCardPaths(t, card,
		"SpellAbility.Val.Modes[0].Sequence[1].Primitive.(game.Pay).Payment.Payer.Val.object.Val.targetIndex = 1",
		"SpellAbility.Val.Modes[0].Sequence[2].Primitive.(game.CounterObject).Object.targetIndex = 1",
	)
}

func TestManaPaymentTargetRemapping(t *testing.T) {
	t.Parallel()
	for _, multiplier := range []bool{false, true} {
		t.Run(map[bool]string{false: "dynamic cost", true: "multiplier"}[multiplier], func(t *testing.T) {
			t.Parallel()
			player := game.TargetPlayerReference(0)
			amount := &game.DynamicAmount{
				Kind:   game.DynamicAmountObjectPower,
				Object: game.TargetPermanentReference(0),
				Player: &player,
				Group:  game.ObjectControlledGroup(game.TargetPermanentReference(0), game.Selection{}),
			}
			payment := game.ResolutionPayment{
				Payer: opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
			}
			if multiplier {
				payment.ManaCostMultiplier = opt.Val(amount)
			} else {
				payment.DynamicGenericManaCost = opt.Val(amount)
			}
			sequence := []game.Instruction{{Primitive: game.Pay{Payment: payment}}}
			if !remapTargetedSequence(sequence, []int{4}) {
				t.Fatal("payment remap failed")
			}
			pay, ok := sequence[0].Primitive.(game.Pay)
			if !ok {
				t.Fatalf("payment primitive = %T", sequence[0].Primitive)
			}
			object, _ := pay.Payment.Payer.Val.Object()
			if object.TargetIndex() != 4 {
				t.Fatalf("payer index = %d, want 4", object.TargetIndex())
			}
			mapped := pay.Payment.DynamicGenericManaCost.Val
			if multiplier {
				mapped = pay.Payment.ManaCostMultiplier.Val
			}
			anchor, _ := mapped.Group.Anchor()
			if mapped.Object.TargetIndex() != 4 || mapped.Player.TargetIndex() != 4 || anchor.TargetIndex() != 4 {
				t.Fatalf("dynamic payment references were not all mapped: %#v", mapped)
			}
			if amount.Object.TargetIndex() != 0 || amount.Player.TargetIndex() != 0 {
				t.Fatal("remapping mutated the original dynamic amount")
			}
		})
	}
}

func TestManaPaymentTargetRemappingRefusals(t *testing.T) {
	t.Parallel()
	for _, payment := range []game.ResolutionPayment{
		{Payer: opt.Val(game.TargetPlayerReference(2))},
		{DynamicGenericManaCost: opt.Val[*game.DynamicAmount](nil)},
		{ManaCostMultiplier: opt.Val[*game.DynamicAmount](nil)},
		{AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalSacrifice}}},
	} {
		sequence := []game.Instruction{{Primitive: game.Pay{Payment: payment}}}
		if remapTargetedSequence(sequence, []int{4}) {
			t.Fatalf("malformed/unmodeled payment remapped: %#v", payment)
		}
	}
}
