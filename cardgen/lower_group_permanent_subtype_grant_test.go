package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerControlledPermanentSubtypeGroupQuotedAbilityGrant verifies that a
// quoted ability granted to a controlled non-creature permanent-subtype group
// ("Foods you control have '<ability>'") lowers to an ability-layer continuous
// effect whose affected group selects the named subtype and grants the
// recursively lowered quoted ability, mirroring the creature-subtype lord form.
func TestLowerControlledPermanentSubtypeGroupQuotedAbilityGrant(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Food Lord",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		ManaCost:   "{3}{G}",
		OracleText: `Foods you control have "{T}: Add {G}."`,
	})
	effect := staticGrantContinuousEffect(t, face)
	if effect.Layer != game.LayerAbility {
		t.Fatalf("layer = %v, want LayerAbility", effect.Layer)
	}
	subtypes := effect.Group.Selection().SubtypesAny
	if len(subtypes) != 1 || subtypes[0] != types.Food {
		t.Fatalf("group subtypes = %#v, want [Food]", subtypes)
	}
	if _, excluded := effect.Group.Exclusion(); excluded {
		t.Fatalf("group = %#v, want no source exclusion for a non-\"other\" group", effect.Group)
	}
	if len(effect.AddAbilities) != 1 {
		t.Fatalf("added abilities = %#v, want exactly one granted ability", effect.AddAbilities)
	}
}

// TestLowerControlledPermanentSubtypeGroupKeywordGrant verifies that a keyword
// granted to a controlled non-creature permanent-subtype group ("Vehicles you
// control have flying") lowers to an ability-layer continuous effect whose group
// selects the named subtype.
func TestLowerControlledPermanentSubtypeGroupKeywordGrant(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Vehicle Lord",
		Layout:     "normal",
		TypeLine:   "Creature — Human Pilot",
		ManaCost:   "{2}{W}",
		OracleText: "Vehicles you control have flying.",
	})
	effect := staticGrantContinuousEffect(t, face)
	subtypes := effect.Group.Selection().SubtypesAny
	if len(subtypes) != 1 || subtypes[0] != types.Vehicle {
		t.Fatalf("group subtypes = %#v, want [Vehicle]", subtypes)
	}
}
