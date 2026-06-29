package cardgen

import (
	"strings"
	"testing"
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
// spell's recipient may be an "another" self-exclusion over a card-type union
// ("another target creature, planeswalker, or battle", Cosmic Hunger): the union
// types lower to RequiredTypesAny and "another" maps to ExcludeSource.
func TestGenerateExecutableCardSourcePowerAnotherUnionRecipient(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Cosmic Bite",
		Layout:     "normal",
		ManaCost:   "{1}{G}",
		TypeLine:   "Instant",
		OracleText: "Target creature you control deals damage equal to its power to another target creature, planeswalker, or battle.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"types.Creature, types.Planeswalker, types.Battle",
		"ExcludeSource: true",
		"game.AnyTargetDamageRecipient(1)",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
