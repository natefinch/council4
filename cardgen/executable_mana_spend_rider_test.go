package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourcePathOfAncestry covers the full Path of
// Ancestry card: a commander-identity mana ability whose produced mana carries a
// one-shot spend rider that scries 1 when the mana is spent to cast a creature
// spell sharing a creature type with the commander.
func TestGenerateExecutableCardSourcePathOfAncestry(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Path of Ancestry",
		Layout:   "normal",
		TypeLine: "Land",
		OracleText: "This land enters tapped.\n" +
			"{T}: Add one mana of any color in your commander's color identity. When that mana is spent to cast a creature spell that shares a creature type with your commander, scry 1. (Look at the top card of your library. You may put that card on the bottom of your library.)",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "p")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"ManaAbilities: []game.ManaAbility",
		"SpendRider: opt.Val(",
		"game.ManaSpendRider{",
		"Condition: game.ManaSpendCastCommanderCreatureType,",
		"game.Scry{",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceCavernOfSouls(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Cavern of Souls",
		Layout:   "normal",
		TypeLine: "Land",
		OracleText: "As this land enters, choose a creature type.\n" +
			"{T}: Add {C}.\n" +
			"{T}: Add one mana of any color. Spend this mana only to cast a creature spell of the chosen type, and that spell can't be countered.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EntryTypeChoiceReplacement(",
		"game.ManaSpendCastChosenCreatureType",
		"game.ManaSpendRestrictedToCondition",
		"game.RuleEffectCantBeCountered",
		"ChosenSubtypeFrom: game.EntryTypeChoiceKey,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceDelightedHalfling covers Delighted Halfling: a
// tap any-color mana ability whose produced mana may be spent only to cast a
// legendary spell, which is additionally made uncounterable. The legendary
// filter is a fixed supertype test, so unlike Cavern of Souls it captures no
// entry-time chosen subtype.
func TestGenerateExecutableCardSourceDelightedHalfling(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Delighted Halfling",
		Layout:   "normal",
		TypeLine: "Creature — Halfling Citizen",
		OracleText: "{T}: Add {C}.\n" +
			"{T}: Add one mana of any color. Spend this mana only to cast a legendary spell, and that spell can't be countered.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "d")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.ManaSpendCastLegendarySpell",
		"game.ManaSpendRestrictedToCondition",
		"game.RuleEffectCantBeCountered",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "ChosenSubtypeFrom:") {
		t.Fatalf("legendary rider must not capture an entry-time subtype:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceChosenTypeManaRiderFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []string{
		"Spend this mana to cast a creature spell of the chosen type, and that spell can't be countered.",
		"Spend this mana only to cast a spell of the chosen type, and that spell can't be countered.",
		"Spend this mana only to cast a creature spell of the chosen type, and that spell cannot be countered.",
		"Spend this mana only to cast a creature spell of the chosen type, and that spell can't be countered by spells.",
	}
	for _, rider := range tests {
		card := &ScryfallCard{
			Name:       "Near Miss Land",
			Layout:     "normal",
			TypeLine:   "Land",
			OracleText: "{T}: Add one mana of any color. " + rider,
		}
		source, diagnostics, err := GenerateExecutableCardSource(card, "n")
		if err != nil {
			t.Fatal(err)
		}
		if len(diagnostics) == 0 {
			t.Fatalf("expected diagnostic for %q, got source:\n%s", rider, source)
		}
		if strings.Contains(source, "game.RuleEffectCantBeCountered") {
			t.Fatalf("near-miss rider %q gained uncounterable semantics:\n%s", rider, source)
		}
	}
}

func TestGenerateExecutableCardSourceChosenTypeManaRiderRequiresEntryChoice(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Choice-Free Cavern",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add one mana of any color. Spend this mana only to cast a creature spell of the chosen type, and that spell can't be countered.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatalf("expected missing entry-choice diagnostic, got source:\n%s", source)
	}
	if strings.Contains(source, "game.RuleEffectCantBeCountered") {
		t.Fatalf("choice-free rider gained executable semantics:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceManaSpendRiderFailsClosed asserts that a
// commander-identity mana ability with a rider the parser does not recognize as
// the exact Path of Ancestry shape (here a different rider effect, "draw a
// card", in place of "scry N") fails closed: it must not lower to a spend-rider
// mana ability and must surface a diagnostic rather than silently dropping the
// rider.
func TestGenerateExecutableCardSourceManaSpendRiderFailsClosed(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Fake Ancestry",
		Layout:   "normal",
		TypeLine: "Land",
		OracleText: "{T}: Add one mana of any color in your commander's color identity. " +
			"When that mana is spent to cast a creature spell that shares a creature type with your commander, draw a card.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatalf("expected fail-closed diagnostic, got source:\n%s", source)
	}
	if strings.Contains(source, "SpendRider: opt.Val(") {
		t.Fatalf("unrecognized rider wrongly lowered to a spend rider:\n%s", source)
	}
}
