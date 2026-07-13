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

// TestGenerateExecutableCardSourceClassBaseLevelAbility verifies a Class whose
// only printed ability is its level-1 base ability (no level-up lines) lowers as
// an ordinary permanent ability with no class-level gate: the base level's
// abilities are always active (CR 716.4), so no SourceClassLevel condition is
// emitted.
func TestGenerateExecutableCardSourceClassBaseLevelAbility(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Base Class",
		Layout:   "class",
		ManaCost: "{G}",
		TypeLine: "Enchantment — Class",
		OracleText: "(Gain the next level as a sorcery to add its ability.)\n" +
			"When this Class enters, create a 2/2 green Wolf creature token.",
	}, "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if strings.Contains(source, "SourceClassLevel") {
		t.Fatalf("base level-1 ability should carry no class-level gate:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceInnkeepersTalentCompiles verifies the full
// Innkeeper's Talent Class lowers end-to-end: its level-1 combat trigger is
// always active, its level-2 filtered ward grant is gated at level 2, and its
// level-3 counter-doubling replacement is gated at level 3 through an ordinary
// source-relative Condition (CR 716). The card generates with no diagnostics.
func TestGenerateExecutableCardSourceInnkeepersTalentCompiles(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Innkeeper's Talent",
		Layout:   "class",
		ManaCost: "{1}{G}",
		TypeLine: "Enchantment — Class",
		OracleText: "(Gain the next level as a sorcery to add its ability.)\n" +
			"At the beginning of combat on your turn, put a +1/+1 counter on target creature you control.\n" +
			"{G}: Level 2\n" +
			"Permanents you control with counters on them have ward {1}.\n" +
			"{3}{G}: Level 3\n" +
			"If you would put one or more counters on a permanent or player, put twice that many of each of those kinds of counters on that permanent or player instead.",
	}, "i")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("expected Innkeeper's Talent to compile, diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		// Level-2 filtered ward grant over every permanent type bearing a counter.
		"game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{MatchAnyCounter: true})",
		"game.WardStaticAbility(cost.Mana{cost.O(1)})",
		"SourceClassLevelAtLeast: 2",
		// Level-3 counter doubler gated by class level through the generic wrapper.
		"game.ClassLevelGatedReplacement(game.AnyCounterPlacementReplacement(",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
