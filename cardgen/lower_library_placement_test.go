package cardgen

import "testing"

// expectLibraryPlacementNotCounter asserts that a "Put <object> on top of/into
// <library>" effect fails closed with an "unsupported library placement"
// diagnostic and is no longer misclassified as "unsupported counter placement".
func expectLibraryPlacementNotCounter(t *testing.T, oracleText string) {
	t.Helper()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Library Placement",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: oracleText,
	})
	for i := range faces {
		if faces[i].SpellAbility.Exists {
			t.Fatalf("%q unexpectedly lowered a spell ability", oracleText)
		}
	}
	library := false
	for i := range diagnostics {
		if diagnostics[i].Summary == "unsupported counter placement" {
			t.Fatalf("%q reported unsupported counter placement, want library placement", oracleText)
		}
		if diagnostics[i].Summary == "unsupported library placement" {
			library = true
		}
	}
	if !library {
		t.Fatalf("diagnostics = %#v, want unsupported library placement", diagnostics)
	}
}

func TestLibraryPlacementOwnerLibraryNotCounter(t *testing.T) {
	t.Parallel()
	// A non-target "put a creature you control on top of its owner's library"
	// (Nulltread Gargantuan) is still an unsupported library placement — the
	// supported forms are the source-permanent tuck (lowerPutSourceOnLibrary) and
	// the single-target tuck (lowerPutTargetOnLibrary) — so it still exercises the
	// owner's-library-vs-counter-placement distinction while remaining unlowered.
	expectLibraryPlacementNotCounter(t, "Put a creature you control on top of its owner's library.")
}

func TestLibraryPlacementTheirLibraryNotCounter(t *testing.T) {
	t.Parallel()
	expectLibraryPlacementNotCounter(
		t,
		"Put up to three target cards from an opponent's graveyard on top of their library in any order.",
	)
}

func TestLibraryPlacementYourLibraryNotCounter(t *testing.T) {
	t.Parallel()
	expectLibraryPlacementNotCounter(
		t,
		"Put target creature card from your graveyard into your library third from the top.",
	)
}
