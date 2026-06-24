package cardgen

import (
	"strings"
	"testing"
)

// TestLowerRosieCottonLegendarySelfReferenceAndExclusion proves Rosie Cotton of
// South Lane lowers fully. It exercises two text-blind features at once: the
// legendary "<First> of <Place>" short-name self-reference ("Rosie Cotton"
// naming the source itself in the enters trigger) and the "other than Rosie
// Cotton" self-exclusion on the +1/+1 counter's single permanent target, which
// lowers to an exclude-source selection.
func TestLowerRosieCottonLegendarySelfReferenceAndExclusion(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:      "Rosie Cotton of South Lane",
		Layout:    "normal",
		TypeLine:  "Legendary Creature — Halfling Peasant",
		ManaCost:  "{2}{W}",
		Power:     new("1"),
		Toughness: new("3"),
		OracleText: "When Rosie Cotton enters, create a Food token. (It's an artifact with \"{2}, {T}, Sacrifice this token: You gain 3 life.\")\n" +
			"Whenever you create a token, put a +1/+1 counter on target creature you control other than Rosie Cotton.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"counter.PlusOnePlusOne",
		"Another:        true,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestLowerRosieCottonShortNameRequiresLegendary proves the multi-word pre-"of"
// short name ("Rosie Cotton") is recognized as a self-reference only when the
// card is legendary. A non-legendary card with the same name and enters trigger
// must not lower the self-named enters trigger, keeping the legendary gate that
// prevents regressions on non-legendary "X of Y" names.
func TestLowerRosieCottonShortNameRequiresLegendary(t *testing.T) {
	t.Parallel()
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Rosie Cotton of South Lane",
		Layout:     "normal",
		TypeLine:   "Creature — Halfling Peasant",
		ManaCost:   "{2}{W}",
		Power:      new("1"),
		Toughness:  new("3"),
		OracleText: "When Rosie Cotton enters, create a Food token.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected the non-legendary short-name enters trigger to remain unsupported")
	}
}
