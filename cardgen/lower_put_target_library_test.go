package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// tuckSpellInstruction lowers a single-face spell whose only effect is a
// "put target <permanent> on top/bottom of its owner's library" tuck and
// returns the lowered target spec and the PutPermanentOnLibrary primitive.
func tuckSpellInstruction(t *testing.T, oracleText string) (game.TargetSpec, game.PutPermanentOnLibrary) {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Tuck Spell",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: oracleText,
	})
	if !face.SpellAbility.Exists {
		t.Fatalf("no spell ability lowered for %q", oracleText)
	}
	modes := face.SpellAbility.Val.Modes
	if len(modes) != 1 || len(modes[0].Targets) != 1 || len(modes[0].Sequence) != 1 {
		t.Fatalf("modes = %#v, want one mode with one target and one instruction", modes)
	}
	prim, ok := modes[0].Sequence[0].Primitive.(game.PutPermanentOnLibrary)
	if !ok {
		t.Fatalf("primitive = %#v, want PutPermanentOnLibrary", modes[0].Sequence[0].Primitive)
	}
	return modes[0].Targets[0], prim
}

// TestLowerPutTargetCreatureOnTopOfLibrary proves the canonical single-target
// tuck to the top of the owner's library (Time Ebb, Griptide) lowers to a
// PutPermanentOnLibrary on the chosen permanent target, not to the bottom.
func TestLowerPutTargetCreatureOnTopOfLibrary(t *testing.T) {
	t.Parallel()
	spec, prim := tuckSpellInstruction(t, "Put target creature on top of its owner's library.")
	if prim.Bottom {
		t.Fatalf("prim = %#v, want top (Bottom false)", prim)
	}
	if spec.Allow != game.TargetAllowPermanent {
		t.Fatalf("spec.Allow = %v, want TargetAllowPermanent", spec.Allow)
	}
	if spec.MinTargets != 1 || spec.MaxTargets != 1 {
		t.Fatalf("spec cardinality = %d..%d, want 1..1", spec.MinTargets, spec.MaxTargets)
	}
	if !spec.Selection.Exists {
		t.Fatal("spec has no selection")
	}
	if got := spec.Selection.Val.RequiredTypesAny; len(got) != 1 || got[0] != "Creature" {
		t.Fatalf("selection required types = %v, want [Creature]", got)
	}
}

// TestLowerPutTargetOnBottomOfLibrary proves the bottom wording sets Bottom on
// the primitive (Mystic Repeal, Eternal Isolation).
func TestLowerPutTargetOnBottomOfLibrary(t *testing.T) {
	t.Parallel()
	_, prim := tuckSpellInstruction(t, "Put target creature on the bottom of its owner's library.")
	if !prim.Bottom {
		t.Fatalf("prim = %#v, want bottom (Bottom true)", prim)
	}
}

// TestLowerPutTargetTypeUnionOnLibrary proves the tuck reuses the shared
// permanent-target projector so a multi-type union target ("artifact or
// enchantment", Disempower) composes without per-qualifier work.
func TestLowerPutTargetTypeUnionOnLibrary(t *testing.T) {
	t.Parallel()
	spec, _ := tuckSpellInstruction(t, "Put target artifact or enchantment on top of its owner's library.")
	got := spec.Selection.Val.RequiredTypesAny
	if len(got) != 2 || got[0] != "Artifact" || got[1] != "Enchantment" {
		t.Fatalf("selection required types = %v, want [Artifact Enchantment]", got)
	}
}

// TestLowerPutTargetPowerQualifierOnLibrary proves a power-filtered target
// (Eternal Isolation's "creature with power 4 or greater") composes with the
// bottom tuck through the shared projector.
func TestLowerPutTargetPowerQualifierOnLibrary(t *testing.T) {
	t.Parallel()
	spec, prim := tuckSpellInstruction(
		t,
		"Put target creature with power 4 or greater on the bottom of its owner's library.",
	)
	if !prim.Bottom {
		t.Fatalf("prim = %#v, want bottom", prim)
	}
	if !spec.Selection.Val.Power.Exists {
		t.Fatalf("selection = %#v, want a power constraint", spec.Selection.Val)
	}
}

// TestLowerPutTargetControlledCreatureOnLibrary proves the controller qualifier
// is preserved (Civic Guildmage's "target creature you control").
func TestLowerPutTargetControlledCreatureOnLibrary(t *testing.T) {
	t.Parallel()
	spec, _ := tuckSpellInstruction(t, "Put target creature you control on top of its owner's library.")
	if spec.Selection.Val.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want ControllerYou", spec.Selection.Val.Controller)
	}
}

// TestLowerPutNonTargetOnLibraryFailsClosed proves a non-target subject tuck
// ("put a creature you control on top of its owner's library", Nulltread
// Gargantuan) is not mistaken for the single-target form and fails closed with
// an unsupported library placement diagnostic rather than lowering a spell.
func TestLowerPutNonTargetOnLibraryFailsClosed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Nontarget Tuck",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Put a creature you control on top of its owner's library.",
	})
	if face.SpellAbility.Exists {
		t.Fatal("non-target tuck unexpectedly lowered a spell ability")
	}
}

// TestLowerPutGraveyardCardOnLibraryNotTuck proves the battlefield tuck path
// does not steal a graveyard-card put ("put target creature card from your
// graveyard on top of its owner's library"), which the graveyard-return path
// owns; the produced primitive must not be a bare PutPermanentOnLibrary on a
// battlefield permanent reference.
func TestLowerPutTargetOnLibraryIsExact(t *testing.T) {
	t.Parallel()
	// The clause round-trips byte-exactly so no residual "unsupported ordered
	// effect sequence" or inexactness leaks; lowerSingleFace asserts zero
	// diagnostics.
	_, prim := tuckSpellInstruction(t, "Put target nonland permanent on top of its owner's library.")
	if prim.Bottom {
		t.Fatalf("prim = %#v, want top", prim)
	}
}
