package parser

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func lexedWords(t *testing.T, source string) []shared.Token {
	t.Helper()
	tokens, diagnostics := lexAll(source)
	if len(diagnostics) != 0 {
		t.Fatalf("lex %q produced diagnostics: %v", source, diagnostics)
	}
	// Drop the trailing EOF token so callers index real tokens directly.
	if len(tokens) > 0 && tokens[len(tokens)-1].Kind == shared.EOF {
		tokens = tokens[:len(tokens)-1]
	}
	return tokens
}

// atomsFor parses a single-ability source and returns the typed atoms the parser
// emitted for it, so tests can assert on the emitted meaning rather than on the
// recognizers used to produce it.
func atomsFor(t *testing.T, source, cardName string) Atoms {
	t.Helper()
	document, diagnostics := Parse(source, Context{CardName: cardName})
	if len(diagnostics) != 0 {
		t.Fatalf("parse %q produced diagnostics: %v", source, diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("parse %q produced %d abilities; want 1", source, len(document.Abilities))
	}
	return document.Abilities[0].Atoms
}

// TestRecognizeColorWordMeaning keeps the parser-owned color vocabulary covered:
// the parser is the single source of color spelling truth.
func TestRecognizeColorWordMeaning(t *testing.T) {
	t.Parallel()
	cases := map[string]Color{
		"white": ColorWhite,
		"WHITE": ColorWhite,
		"blue":  ColorBlue,
		"black": ColorBlack,
		"red":   ColorRed,
		"Green": ColorGreen,
	}
	for word, want := range cases {
		got, ok := recognizeColorWord(word)
		if !ok || got != want {
			t.Errorf("recognizeColorWord(%q) = %v, %v; want %v, true", word, got, ok, want)
		}
	}
	for _, word := range []string{"colorless", "multicolored", "gold", "whitish", ""} {
		if got, ok := recognizeColorWord(word); ok {
			t.Errorf("recognizeColorWord(%q) = %v, true; want unknown, false", word, got)
		}
	}
}

// TestParseEmitsColorAtoms proves Parse emits a source-spanned color atom for
// each color word and that the span covers exactly that word.
func TestParseEmitsColorAtoms(t *testing.T) {
	t.Parallel()
	source := "destroy all white and black creatures"
	atoms := atomsFor(t, source, "")
	if len(atoms.Colors()) != 2 {
		t.Fatalf("emitted %d color atoms; want 2 (%+v)", len(atoms.Colors()), atoms.Colors())
	}
	if got := shared.SliceSpan(source, atoms.Colors()[0].Span); got != "white" {
		t.Errorf("first color span = %q; want %q", got, "white")
	}
	if atoms.Colors()[0].Color != ColorWhite || atoms.Colors()[1].Color != ColorBlack {
		t.Errorf("color atoms = %v; want white then black", atoms.Colors())
	}
	// A color atom is looked up by the span it begins at.
	if color, ok := atoms.ColorAt(atoms.Colors()[1].Span); !ok || color != ColorBlack {
		t.Errorf("ColorAt(black span) = %v, %v; want black, true", color, ok)
	}
	// Near miss: "colorless" is not a color and is not emitted.
	if len(atomsFor(t, "destroy all colorless permanents", "").Colors()) != 0 {
		t.Error("emitted a color atom for colorless; want none")
	}
	if len(atomsFor(t, "destroy target nonblack creature", "").Colors()) != 0 {
		t.Error("emitted a positive color atom for nonblack; want none")
	}
	nonblack := atomsFor(t, "destroy target nonblack creature", "")
	if len(nonblack.ExcludedColors()) != 1 {
		t.Fatalf("excluded color atoms = %+v; want one", nonblack.ExcludedColors())
	}
	if color, ok := nonblack.ExcludedColorAt(nonblack.ExcludedColors()[0].Span); !ok || color != ColorBlack {
		t.Errorf("ExcludedColorAt(nonblack) = %v, %v; want black, true", color, ok)
	}
}

// TestParseEmitsCardTypeAtoms proves card-type atoms are emitted with plural
// normalization owned by the parser.
func TestParseEmitsCardTypeAtoms(t *testing.T) {
	t.Parallel()
	atoms := atomsFor(t, "destroy target artifact or enchantments", "")
	got := make([]CardType, len(atoms.CardTypes()))
	for i, atom := range atoms.CardTypes() {
		got[i] = atom.Type
	}
	if !slices.Equal(got, []CardType{CardTypeArtifact, CardTypeEnchantment}) {
		t.Fatalf("card type atoms = %v; want artifact then enchantment", got)
	}
	// "permanent" is not a card type and fails closed.
	if len(atomsFor(t, "exile target permanent", "").CardTypes()) != 0 {
		t.Error("emitted a card-type atom for permanent; want none")
	}
	noncreature := atomsFor(t, "counter target noncreature spell", "")
	if len(noncreature.CardTypes()) != 0 {
		t.Errorf("emitted positive card-type atoms for noncreature: %+v", noncreature.CardTypes())
	}
	if len(noncreature.ExcludedTypes()) != 1 ||
		noncreature.ExcludedTypes()[0].Type != CardTypeCreature {
		t.Errorf("excluded card-type atoms = %+v; want creature", noncreature.ExcludedTypes())
	}
}

// TestParseEmitsSubtypeAtoms proves subtype identities across card-type families
// are emitted with plural, multiword, and capitalization normalization, failing
// closed otherwise.
func TestParseEmitsSubtypeAtoms(t *testing.T) {
	t.Parallel()
	cases := map[string]types.Sub{
		"protection from Dragons":           types.Dragon,
		"destroy all Elves":                 types.Elf,
		"exile target Clues":                types.Clue,
		"destroy target Auras":              types.Aura,
		"return target Power Plants":        types.PowerPlant,
		"counter target Arcane spell":       types.Arcane,
		"whenever a Time Lord dies":         types.TimeLord,
		"destroy target Sieges":             types.Siege,
		"whenever a planeswalker Jace dies": types.Jace,
	}
	for source, want := range cases {
		atoms := atomsFor(t, source, "")
		if got, ok := lastSubtype(atoms); !ok || got != want {
			t.Errorf("%q emitted subtype %v, %v; want %v", source, got, ok, want)
		}
	}
	// A non-subtype noun fails closed.
	for _, atom := range atomsFor(t, "draw a card", "").Subtypes() {
		t.Errorf("emitted unexpected subtype atom %v", atom)
	}
}

func lastSubtype(atoms Atoms) (types.Sub, bool) {
	if len(atoms.Subtypes()) == 0 {
		return "", false
	}
	return atoms.Subtypes()[len(atoms.Subtypes())-1].Identity, true
}

// TestParseEmitsZoneAtoms proves zone atoms carry the typed runtime zone and the
// role their introducing wording gives them.
func TestParseEmitsZoneAtoms(t *testing.T) {
	t.Parallel()
	atoms := atomsFor(t, "return target creature card from your graveyard to your hand", "")
	if z, ok := zoneRole(atoms, ZoneRoleFrom); !ok || z != zone.Graveyard {
		t.Errorf("from zone = %v, %v; want graveyard, true", z, ok)
	}
	if z, ok := zoneRole(atoms, ZoneRoleTo); !ok || z != zone.Hand {
		t.Errorf("to zone = %v, %v; want hand, true", z, ok)
	}
	// "draw a card" introduces no zone movement.
	if len(atomsFor(t, "draw a card", "").Zones()) != 0 {
		t.Error("emitted a zone atom for a zoneless effect; want none")
	}
}

func zoneRole(atoms Atoms, role ZoneRole) (zone.Type, bool) {
	for _, atom := range atoms.Zones() {
		if atom.Role == role {
			return atom.Zone, true
		}
	}
	return zone.None, false
}

// TestParseEmitsCounterAtoms proves a counter atom carries the typed counter kind
// and spans the counter-kind name preceding the "counter" noun.
func TestParseEmitsCounterAtoms(t *testing.T) {
	t.Parallel()
	source := "put two +1/+1 counters on target creature"
	atoms := atomsFor(t, source, "")
	if len(atoms.Counters()) != 1 {
		t.Fatalf("emitted %d counter atoms; want 1 (%+v)", len(atoms.Counters()), atoms.Counters())
	}
	if atoms.Counters()[0].Kind != counter.PlusOnePlusOne {
		t.Errorf("counter kind = %v; want +1/+1", atoms.Counters()[0].Kind)
	}
	if got := shared.SliceSpan(source, atoms.Counters()[0].Span); got != "+1/+1" {
		t.Errorf("counter name span = %q; want %q", got, "+1/+1")
	}
	ageSource := "put an age counter on this permanent"
	ageAtoms := atomsFor(t, ageSource, "")
	if len(ageAtoms.Counters()) != 1 ||
		ageAtoms.Counters()[0].Kind != counter.Age ||
		shared.SliceSpan(ageSource, ageAtoms.Counters()[0].Span) != "age" {
		t.Errorf("age counter atoms = %+v; want one source-spanned age counter", ageAtoms.Counters())
	}
	// A bare "draw a card" emits no counter atom.
	if len(atomsFor(t, "draw a card", "").Counters()) != 0 {
		t.Error("emitted a counter atom where none exists; want none")
	}
}

// TestRecognizeCardinalVocabulary keeps the parser-owned cardinal vocabulary
// covered and asserts the emitted typed values.
func TestRecognizeCardinalVocabulary(t *testing.T) {
	t.Parallel()
	cases := map[string]int{
		"one": 1, "two": 2, "three": 3, "four": 4, "five": 5,
		"six": 6, "seven": 7, "eight": 8, "nine": 9, "ten": 10, "TEN": 10,
	}
	for word, want := range cases {
		if got, ok := CardinalWordValue(word); !ok || got != want {
			t.Errorf("CardinalWordValue(%q) = %v, %v; want %v, true", word, got, ok, want)
		}
	}
	for _, word := range []string{"eleven", "zero", "1", "", "twenty"} {
		if got, ok := CardinalWordValue(word); ok {
			t.Errorf("CardinalWordValue(%q) = %v, true; want 0, false", word, got)
		}
	}
	atoms := atomsFor(t, "draw three cards", "")
	if len(atoms.Cardinals()) != 1 || atoms.Cardinals()[0].Value != 3 {
		t.Fatalf("cardinal atoms = %+v; want one atom valued 3", atoms.Cardinals())
	}
}

func TestSingularNounForms(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"Elves":    "Elf",
		"Pixies":   "Pixy",
		"Wolves":   "Wolf",
		"Goblins":  "Goblin",
		"Wall":     "Wall",
		"Sphinxes": "Sphinx",
	}
	for plural, wantSingular := range cases {
		forms := SingularNounForms(plural)
		if forms[0] != plural {
			t.Errorf("SingularNounForms(%q)[0] = %q; want input first", plural, forms[0])
		}
		if !slices.Contains(forms, wantSingular) {
			t.Errorf("SingularNounForms(%q) = %v; want to contain %q", plural, forms, wantSingular)
		}
	}
}

func TestParseEmitsExtendedAtomVocabulary(t *testing.T) {
	t.Parallel()
	atoms := atomsFor(t, "Whenever another nontoken multicolored Aura you control enters from exile, put a stun counter on it.", "Test Aura")
	if len(atoms.ColorQualifiers()) == 0 || atoms.ColorQualifiers()[0].Qualifier != ColorQualifierMulticolored {
		t.Fatalf("color qualifiers = %+v; want multicolored", atoms.ColorQualifiers())
	}
	if !hasSelectionFlag(atoms, SelectionFlagAnother) || !hasSelectionFlag(atoms, SelectionFlagNonToken) {
		t.Fatalf("selection flags = %+v; want another and nontoken", atoms.SelectionFlags())
	}
	if relation, ok := atoms.ControllerIn(atoms.Controllers()[0].Span); !ok || relation != ControllerRelationYouControl {
		t.Fatalf("controller relation = %v, %v; want you-control", relation, ok)
	}
	if z, ok := zoneRole(atoms, ZoneRoleFrom); !ok || z != zone.Exile {
		t.Fatalf("from zone = %v, %v; want exile", z, ok)
	}
}

func TestParseEmitsOrdinalAtoms(t *testing.T) {
	t.Parallel()
	atoms := atomsFor(t, "Whenever you draw your third card each turn, draw a card.", "")
	if len(atoms.Ordinals()) != 1 || atoms.Ordinals()[0].Value != 3 {
		t.Fatalf("ordinals = %+v; want third=3", atoms.Ordinals())
	}
	if value, ok := OrdinalWordValue("sixth"); ok || value != 0 {
		t.Fatalf("OrdinalWordValue(sixth) = %d, %v; want 0, false", value, ok)
	}
}

func TestParseSelfNameAliasesDoNotDuplicateLongName(t *testing.T) {
	t.Parallel()
	atoms := atomsFor(t, "Dragon's Rage Channeler gets +2/+2.", "Dragon's Rage Channeler")
	if len(atoms.SelfNameSpans()) != 1 {
		t.Fatalf("self-name spans = %+v; want one full-name span", atoms.SelfNameSpans())
	}
	atoms = atomsFor(t, "Yomiji dies.", "Yomiji, Who Bars the Way")
	if len(atoms.SelfNameSpans()) != 1 {
		t.Fatalf("short self-name spans = %+v; want one", atoms.SelfNameSpans())
	}
}

func hasSelectionFlag(atoms Atoms, flag SelectionFlag) bool {
	for _, atom := range atoms.SelectionFlags() {
		if atom.Flag == flag {
			return true
		}
	}
	return false
}
