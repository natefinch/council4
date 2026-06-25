package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceClassLevelUp verifies the Class level-up slice:
// each "{cost}: Level N" line lowers to a sorcery-timed activated ability that
// sets the source's class level, gated so it can only raise the level by one,
// and abilities printed after a level line are gated to that class level.
func TestGenerateExecutableCardSourceClassLevelUp(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Hunter's Talent",
		Layout:   "class",
		ManaCost: "{1}{G}",
		TypeLine: "Enchantment — Class",
		OracleText: "(Gain the next level as a sorcery to add its ability.)\n" +
			"When this Class enters, target creature you control deals damage equal to its power to target creature you don't control.\n" +
			"{1}{G}: Level 2\n" +
			"Whenever you attack, target attacking creature gets +1/+0 and gains trample until end of turn.\n" +
			"{3}{G}: Level 3\n" +
			"At the beginning of your end step, if you control a creature with power 4 or greater, draw a card.",
	}, "h")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.SetClassLevel{",
		"game.SorceryOnly",
		"SourceClassLevelLessThan: 2",
		"SourceClassLevelLessThan: 3",
		"SourceClassLevelAtLeast: 2",
		"SourceClassLevelAtLeast: 3",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

// TestGenerateExecutableCardSourceClassBecameLevel verifies the "When this Class
// becomes level N" trigger lowers to a class-level-gained event pattern
// restricted to the level reached, gated by the level band it sits in.
func TestGenerateExecutableCardSourceClassBecameLevel(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Wizard Class",
		Layout:   "class",
		ManaCost: "{U}",
		TypeLine: "Enchantment — Class",
		OracleText: "(Gain the next level as a sorcery to add its ability.)\n" +
			"You have no maximum hand size.\n" +
			"{2}{U}: Level 2\n" +
			"When this Class becomes level 2, draw two cards.\n" +
			"{4}{U}: Level 3\n" +
			"Whenever you draw a card, put a +1/+1 counter on target creature you control.",
	}, "w")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.EventClassLevelGained",
		"ClassBecameLevel: 2",
		"game.TriggerSourceSelf",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
