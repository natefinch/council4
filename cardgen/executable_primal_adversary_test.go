package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutablePrimalAdversary exercises the repeatable-payment land
// animation Primal Adversary combines (issue #1564): an enters trigger that
// offers {1}{G} any number of times, publishes the payment count, puts that many
// +1/+1 counters on the source, and lets the controller animate up to that many
// lands they control into 3/3 Wolf creatures with haste that remain lands. The
// PayRepeatedly count drives both the AddCounter amount and the ApplyContinuous
// choose-up-to land selection through the same DynamicAmountChosenNumber key.
func TestGenerateExecutablePrimalAdversary(t *testing.T) {
	t.Parallel()
	power := "4"
	toughness := "3"
	card := &ScryfallCard{
		Name:     "Primal Adversary",
		Layout:   "normal",
		ManaCost: "{2}{G}",
		TypeLine: "Creature — Wolf",
		OracleText: "Trample\n" +
			"When this creature enters, you may pay {1}{G} any number of times. " +
			"When you pay this cost one or more times, put that many +1/+1 counters on this creature, " +
			"then up to that many target lands you control become 3/3 Wolf creatures with haste that are still lands.",
		Colors:    []string{"G"},
		Power:     &power,
		Toughness: &toughness,
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.PayRepeatedly{",
		`PublishCount: "pay-repeatedly-count",`,
		"Primitive: game.AddCounter{",
		"Kind:      game.DynamicAmountChosenNumber,",
		`ResultKey: game.ResultKey("pay-repeatedly-count"),`,
		"CounterKind: counter.PlusOnePlusOne,",
		"Primitive: game.ApplyContinuous{",
		"ChooseFrom: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Land}}),",
		"AddTypes:    []types.Card{types.Creature},",
		"AddSubtypes: []types.Sub{types.Wolf},",
		"game.Haste,",
		"SetPower:     opt.Val(game.PT{Value: 3}),",
		"Duration:   game.DurationPermanent,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
