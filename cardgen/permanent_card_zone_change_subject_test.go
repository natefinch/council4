package cardgen

import (
	"strings"
	"testing"
)

// TestPermanentCardZoneChangeSubjectTypesInNonBattlefieldSource verifies that a
// "permanent card" subject on a zone-change trigger whose source is a
// non-battlefield zone (here the library) types to the permanent-card union so
// instant and sorcery cards are excluded. A library holds every card type, so
// the restriction is meaningful and must be emitted.
func TestPermanentCardZoneChangeSubjectTypesInNonBattlefieldSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Library Permanent Trigger",
		Layout:     "normal",
		ManaCost:   "{1}",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a permanent card is put into your graveyard from your library, you gain 1 life.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	want := "SubjectSelection: game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Battle, types.Creature, types.Enchantment, types.Land, types.Planeswalker}}"
	if !strings.Contains(source, want) {
		t.Fatalf("source missing permanent-card-union subject %q:\n%s", want, source)
	}
}

// TestPermanentCardZoneChangeSubjectUntypedInBattlefieldSource verifies that a
// "permanent" subject on a battlefield-origin zone change is left as an any-card
// selection: only permanents exist on the battlefield, so the permanent-card
// restriction is redundant and must not be emitted (preserving prior output).
func TestPermanentCardZoneChangeSubjectUntypedInBattlefieldSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Battlefield Permanent Trigger",
		Layout:     "normal",
		ManaCost:   "{1}",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a permanent you control leaves the battlefield, you gain 1 life.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if strings.Contains(source, "RequiredTypesAny") {
		t.Fatalf("battlefield-origin permanent subject unexpectedly typed its selection:\n%s", source)
	}
}
