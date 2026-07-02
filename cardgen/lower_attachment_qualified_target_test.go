package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// spellDestroyTarget lowers a single-face instant whose only effect is a
// destroy and returns the lowered target's selection.
func spellDestroyTarget(t *testing.T, oracleText string) game.Selection {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Destroy Spell",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: oracleText,
	})
	if !face.SpellAbility.Exists {
		t.Fatalf("no spell ability lowered for %q", oracleText)
	}
	modes := face.SpellAbility.Val.Modes
	if len(modes) != 1 || len(modes[0].Targets) != 1 {
		t.Fatalf("modes = %#v, want one mode with one target", modes)
	}
	target := modes[0].Targets[0]
	if !target.Selection.Exists {
		t.Fatalf("target has no selection for %q", oracleText)
	}
	return target.Selection.Val
}

// TestLowerEnchantedPermanentTargetCarriesMatchEnchanted verifies that a
// destroy on "target enchanted permanent" lowers with the runtime
// MatchEnchanted predicate set (Venomous Vines).
func TestLowerEnchantedPermanentTargetCarriesMatchEnchanted(t *testing.T) {
	t.Parallel()
	selection := spellDestroyTarget(t, "Destroy target enchanted permanent.")
	if !selection.MatchEnchanted {
		t.Fatalf("selection = %#v, want MatchEnchanted", selection)
	}
	if selection.MatchModified || selection.MatchEquipped {
		t.Fatalf("selection = %#v, want only MatchEnchanted", selection)
	}
}

// TestLowerEnchantedCreatureTargetCarriesMatchEnchanted verifies the creature
// noun form also carries MatchEnchanted (Ramses Overdark's ability shape).
func TestLowerEnchantedCreatureTargetCarriesMatchEnchanted(t *testing.T) {
	t.Parallel()
	selection := spellDestroyTarget(t, "Destroy target enchanted creature.")
	if !selection.MatchEnchanted {
		t.Fatalf("selection = %#v, want MatchEnchanted", selection)
	}
}

// TestLowerModifiedCreatureTargetCarriesMatchModified verifies that a destroy
// on "target modified creature you control" lowers with MatchModified and the
// controller restriction preserved.
func TestLowerModifiedCreatureTargetCarriesMatchModified(t *testing.T) {
	t.Parallel()
	selection := spellDestroyTarget(t, "Destroy target modified creature you control.")
	if !selection.MatchModified {
		t.Fatalf("selection = %#v, want MatchModified", selection)
	}
	if selection.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want ControllerYou", selection.Controller)
	}
}

// TestLowerEquippedCreatureTargetCarriesMatchEquipped verifies the equipped
// adjective lowers with MatchEquipped.
func TestLowerEquippedCreatureTargetCarriesMatchEquipped(t *testing.T) {
	t.Parallel()
	selection := spellDestroyTarget(t, "Destroy target equipped creature.")
	if !selection.MatchEquipped {
		t.Fatalf("selection = %#v, want MatchEquipped", selection)
	}
}
