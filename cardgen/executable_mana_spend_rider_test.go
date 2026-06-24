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

// TestGenerateExecutableCardSourceArenaOfGlory covers the unrestricted
// creature-spell haste bonus rider: the exert mana ability adds {R}{R}, and a
// creature spell paid for with that mana gains haste until end of turn. Both
// produced red units carry the rider, the condition is unrestricted, and the
// granted keyword is haste.
func TestGenerateExecutableCardSourceArenaOfGlory(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Arena of Glory",
		Layout:   "normal",
		TypeLine: "Land",
		OracleText: "This land enters tapped unless you control a Mountain.\n" +
			"{T}: Add {R}.\n" +
			"{R}, {T}, Exert this land: Add {R}{R}. If that mana is spent on a creature spell, it gains haste until end of turn.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.ManaSpendCastCreatureSpell",
		"SpellGainsKeywords: []game.Keyword{",
		"game.Haste,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "game.ManaSpendRestrictedToCondition") {
		t.Fatalf("creature-spell haste rider must be unrestricted:\n%s", source)
	}
	if strings.Count(source, "game.ManaSpendCastCreatureSpell") != 2 {
		t.Fatalf("both produced red units must carry the rider:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceSecludedCourtyard covers the cast-or-activate
// chosen-type restriction: the produced mana may be spent to cast a creature
// spell of the chosen type or to activate an ability of a creature source of the
// chosen type. It captures the entry-time chosen subtype and applies no spell
// rule effect.
func TestGenerateExecutableCardSourceSecludedCourtyard(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Secluded Courtyard",
		Layout:   "normal",
		TypeLine: "Land",
		OracleText: "As this land enters, choose a creature type.\n" +
			"{T}: Add {C}.\n" +
			"{T}: Add one mana of any color. Spend this mana only to cast a creature spell of the chosen type or activate an ability of a creature source of the chosen type.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "s")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.EntryTypeChoiceReplacement(",
		"game.ManaSpendCastOrActivateChosenCreatureType",
		"game.ManaSpendRestrictedToCondition",
		"ChosenSubtypeFrom: game.EntryTypeChoiceKey,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "game.RuleEffectCantBeCountered") {
		t.Fatalf("cast-or-activate rider must not make spells uncounterable:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceUnclaimedTerritory covers the bare chosen-type
// restriction (also Pillar of Origins): the produced mana may be spent only to
// cast a creature spell of the chosen type, with no can't-be-countered clause.
func TestGenerateExecutableCardSourceUnclaimedTerritory(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Unclaimed Territory",
		Layout:   "normal",
		TypeLine: "Land",
		OracleText: "As this land enters, choose a creature type.\n" +
			"{T}: Add {C}.\n" +
			"{T}: Add one mana of any color. Spend this mana only to cast a creature spell of the chosen type.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "u")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.ManaSpendCastChosenCreatureType",
		"game.ManaSpendRestrictedToCondition",
		"ChosenSubtypeFrom: game.EntryTypeChoiceKey,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "game.RuleEffectCantBeCountered") {
		t.Fatalf("bare chosen-type rider must not make spells uncounterable:\n%s", source)
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

// TestGenerateExecutableCardSourceBeastcallerSavant covers Beastcaller Savant: a
// tap-for-any-color mana ability whose produced mana may be spent only to cast a
// creature spell. The bare restricted creature-spell rider carries no further
// qualifier or rule effect.
func TestGenerateExecutableCardSourceBeastcallerSavant(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Beastcaller Savant",
		Layout:     "normal",
		TypeLine:   "Creature — Elf Shaman Ally",
		ManaCost:   "{1}{G}",
		Colors:     []string{"G"},
		Power:      new("1"),
		Toughness:  new("1"),
		OracleText: "Haste\n{T}: Add one mana of any color. Spend this mana only to cast a creature spell.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "b")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"SpendRider: opt.Val(",
		"game.ManaSpendCastCreatureSpell",
		"game.ManaSpendRestrictedToCondition",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
	if strings.Contains(source, "SpellGainsKeywords") {
		t.Fatalf("restricted creature-spell rider must not grant keywords:\n%s", source)
	}
}

// TestGenerateExecutableCardSourceFixedColorCreatureSpellRider covers a
// fixed-color add-mana producer (Dwynen's Elite style) carrying the restricted
// creature-spell spend rider, exercising the attach-rider path rather than the
// any-color choice helper.
func TestGenerateExecutableCardSourceFixedColorCreatureSpellRider(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Test Restricted Rock",
		Layout:     "normal",
		TypeLine:   "Artifact",
		OracleText: "{T}: Add {G}. Spend this mana only to cast a creature spell.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"SpendRider: opt.Val(",
		"game.ManaSpendCastCreatureSpell",
		"game.ManaSpendRestrictedToCondition",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceInvigoratingHotSpring covers the conjoined
// "Activate only as a sorcery and only once each turn." timing restriction on a
// counter-removal activated ability.
func TestGenerateExecutableCardSourceInvigoratingHotSpring(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Invigorating Hot Spring",
		Layout:   "normal",
		TypeLine: "Enchantment",
		ManaCost: "{1}{R}{G}",
		Colors:   []string{"R", "G"},
		OracleText: "This enchantment enters with four +1/+1 counters on it.\n" +
			"Modified creatures you control have haste. (Equipment, Auras you control, and counters are modifications.)\n" +
			"Remove a +1/+1 counter from this enchantment: Put a +1/+1 counter on target creature you control. Activate only as a sorcery and only once each turn.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "i")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.SorceryOncePerTurn") {
		t.Fatalf("source missing conjoined sorcery-once-per-turn timing:\n%s", source)
	}
}
