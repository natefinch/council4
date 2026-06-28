package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceCounterCountDamageActivated proves a single
// permanent's "deals damage equal to the number of <kind> counters on it" damage
// ability lowers to the DynamicAmountObjectCounters amount reading the source
// permanent's counters, the counter-count sibling of the source-power damage
// staple (Spikeshot Goblin, Ghitu Fire-Eater). Magma Mine's sacrifice ability
// deals damage to any target equal to its own pressure counters.
func TestGenerateExecutableCardSourceCounterCountDamageActivated(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Test Pressure Mine",
		Layout:   "normal",
		TypeLine: "Artifact",
		OracleText: "{4}: Put a pressure counter on this artifact.\n" +
			"{T}, Sacrifice this artifact: It deals damage equal to the number of pressure counters on it to any target.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.Damage{",
		"game.AnyTargetDamageRecipient(0)",
		"Kind:        game.DynamicAmountObjectCounters",
		"CounterKind: counter.Pressure",
		"Object:      game.SourcePermanentReference()",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

// TestGenerateExecutableCardSourceCounterCountDamageTrigger proves the attack
// trigger form ("Whenever this creature attacks, it deals damage to any target
// equal to the number of +1/+1 counters on this creature.", Preyseizer Dragon)
// reads the triggering permanent's counters via DynamicAmountObjectCounters bound
// to the event permanent.
func TestGenerateExecutableCardSourceCounterCountDamageTrigger(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Counter Dragon",
		Layout:     "normal",
		ManaCost:   "{4}{R}{R}",
		TypeLine:   "Creature — Dragon",
		OracleText: "Whenever this creature attacks, it deals damage to any target equal to the number of +1/+1 counters on this creature.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.Damage{",
		"game.AnyTargetDamageRecipient(0)",
		"Kind:        game.DynamicAmountObjectCounters",
		"CounterKind: counter.PlusOnePlusOne",
		"Object:      game.SourcePermanentReference()",
		"DamageSource: opt.Val(game.EventPermanentReference())",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

// TestGenerateExecutableCardSourceCounterCountDamageWhereX proves the "where X is
// the number of <kind> counters on <self>" variant (Torch Song) lowers the same
// way as the "equal to" form, and that a target-creature recipient (rather than
// any target) is honored.
func TestGenerateExecutableCardSourceCounterCountDamageWhereX(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Verse Burn",
		Layout:     "normal",
		ManaCost:   "{2}{R}",
		TypeLine:   "Enchantment",
		OracleText: "{2}{R}, Sacrifice this enchantment: It deals X damage to target creature, where X is the number of verse counters on this enchantment.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.Damage{",
		"Kind:        game.DynamicAmountObjectCounters",
		"CounterKind: counter.Verse",
		"Object:      game.SourcePermanentReference()",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}

// TestGenerateExecutableCardSourceLeadingAddendCountDamage proves the leading
// addend form "where X is N plus the number of <count>" (Welding Sparks) lowers
// to a DynamicAmountCountSelector that carries the constant N as the amount's
// Addend, the dynamic sibling of the trailing "<count> plus N" addend.
func TestGenerateExecutableCardSourceLeadingAddendCountDamage(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Welding Sparks",
		Layout:     "normal",
		ManaCost:   "{2}{R}",
		TypeLine:   "Instant",
		OracleText: "Test Welding Sparks deals X damage to target creature, where X is 3 plus the number of artifacts you control.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.Damage{",
		"game.AnyTargetDamageRecipient(0)",
		"Kind:       game.DynamicAmountCountSelector",
		"Addend:     3",
		"RequiredTypes: []types.Card{types.Artifact}",
		"Controller: game.ControllerYou",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
