package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestGenerateExecutableCardSourcePowerEachOfRecipients proves a one-sided
// source-power bite spell whose recipient is a plural "each of N other target
// creatures" slot (Betrayal at the Vault) unrolls one power-scaled Damage per
// recipient slot, keyed past the single dealing target. The dealer is the first
// target and its power feeds every instruction.
func TestGenerateExecutableCardSourcePowerEachOfRecipients(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Vault Betrayal",
		Layout:     "normal",
		ManaCost:   "{4}{G}{G}",
		TypeLine:   "Instant",
		OracleText: "Target creature you control deals damage equal to its power to each of two other target creatures.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"MinTargets: 2,",
		"MaxTargets: 2,",
		"game.AnyTargetDamageRecipient(1)",
		"game.AnyTargetDamageRecipient(2)",
		"Kind:       game.DynamicAmountObjectPower",
		"Object:     game.TargetPermanentReference(0)",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

// TestGenerateExecutableCardSourcePowerAnotherUnionRecipient proves the bite
// spell's recipient may be an "another" target over a card-type union
// ("another target creature, planeswalker, or battle", Cosmic Hunger): the union
// types lower to RequiredTypesAny and "another" excludes the prior dealing target.
func TestGenerateExecutableCardSourcePowerAnotherUnionRecipient(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Cosmic Bite",
		Layout:     "normal",
		ManaCost:   "{1}{G}",
		TypeLine:   "Instant",
		OracleText: "Target creature you control deals damage equal to its power to another target creature, planeswalker, or battle.",
	}
	face := lowerSingleFace(t, card)
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 2 || mode.Targets[0].DistinctFromPriorTargets || !mode.Targets[1].DistinctFromPriorTargets {
		t.Fatalf("targets = %+v, want distinct recipient, not distinct dealer", mode.Targets)
	}
	selection := mode.Targets[1].Selection
	if !selection.Exists || selection.Val.ExcludeSource ||
		len(selection.Val.RequiredTypesAny) != 3 ||
		selection.Val.RequiredTypesAny[0] != types.Creature ||
		selection.Val.RequiredTypesAny[1] != types.Planeswalker ||
		selection.Val.RequiredTypesAny[2] != types.Battle {
		t.Fatalf("recipient selection = %+v, want type union without spell-source exclusion", selection)
	}
	damage, ok := mode.Sequence[0].Primitive.(game.Damage)
	if !ok || damage.Recipient != game.AnyTargetDamageRecipient(1) {
		t.Fatalf("damage = %+v, want second target recipient", mode.Sequence[0].Primitive)
	}
	dynamic := damage.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountObjectPower ||
		dynamic.Val.Object != game.TargetPermanentReference(0) {
		t.Fatalf("damage amount = %+v, want dealing target power", dynamic)
	}
}
