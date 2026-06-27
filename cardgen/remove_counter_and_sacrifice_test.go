package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceRemoveCounterAndSacrifice covers the
// Quest/Expedition family's combined "Remove N <kind> counters from this and
// sacrifice it" activation cost. The single cost phrase lowers into two typed
// additional costs: a counter removal from the source and a sacrifice of the
// source. The trailing "it" denotes the source even though the bare pronoun is
// not otherwise bound to it.
func TestGenerateExecutableCardSourceRemoveCounterAndSacrifice(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Test Quest",
		Layout:   "normal",
		TypeLine: "Enchantment",
		ManaCost: "{B}",
		OracleText: "Whenever a creature dies, you may put a quest counter on this enchantment.\n" +
			"Remove three quest counters from this enchantment and sacrifice it: Create a 5/5 black Zombie Giant creature token.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Kind:        cost.AdditionalRemoveCounter",
		"CounterKind: counter.Quest",
		"Kind:   cost.AdditionalSacrificeSource",
	} {
		if !strings.Contains(normalizeSource(source), normalizeSource(wanted)) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
