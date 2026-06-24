package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// staticGrantContinuousEffect returns the single continuous effect of a face
// whose only ability is a static quoted-ability grant to a battlefield group.
func staticGrantContinuousEffect(t *testing.T, face loweredFaceAbilities) game.ContinuousEffect {
	t.Helper()
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %#v, want exactly one", face.StaticAbilities)
	}
	effects := face.StaticAbilities[0].Body.ContinuousEffects
	if len(effects) != 1 {
		t.Fatalf("continuous effects = %#v, want exactly one", effects)
	}
	return effects[0]
}

// TestLowerCreatureSubtypeGroupQuotedAbilityGrant verifies that a tribal lord
// granting a quoted ability to a creature-subtype group ("Sliver creatures you
// control have '<ability>'") lowers to an ability-layer continuous effect whose
// affected group selects the named subtype and whose granted ability is the
// recursively lowered quoted ability.
func TestLowerCreatureSubtypeGroupQuotedAbilityGrant(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Sliver Lord",
		Layout:     "normal",
		TypeLine:   "Creature — Sliver",
		ManaCost:   "{1}{G}",
		OracleText: `Sliver creatures you control have "Whenever this creature attacks, it deals 1 damage to any target."`,
	})
	effect := staticGrantContinuousEffect(t, face)
	if effect.Layer != game.LayerAbility {
		t.Fatalf("layer = %v, want LayerAbility", effect.Layer)
	}
	subtypes := effect.Group.Selection().SubtypesAny
	if len(subtypes) != 1 || subtypes[0] != types.Sliver {
		t.Fatalf("group subtypes = %#v, want [Sliver]", subtypes)
	}
	if _, excluded := effect.Group.Exclusion(); excluded {
		t.Fatalf("group = %#v, want no source exclusion for a non-\"other\" lord", effect.Group)
	}
	if len(effect.AddAbilities) != 1 {
		t.Fatalf("added abilities = %#v, want exactly one granted ability", effect.AddAbilities)
	}
	if _, ok := effect.AddAbilities[0].(*game.TriggeredAbility); !ok {
		t.Fatalf("granted ability = %#v, want a triggered ability", effect.AddAbilities[0])
	}
}

// TestLowerOtherCreatureSubtypeGroupQuotedAbilityGrant verifies that the
// source-excluding "Other <Subtype>s you control have '<ability>'" form lowers
// to a subtype-filtered group that excludes the source permanent and grants the
// quoted activated ability.
func TestLowerOtherCreatureSubtypeGroupQuotedAbilityGrant(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Wizard Lord",
		Layout:     "normal",
		TypeLine:   "Creature — Vedalken Wizard",
		ManaCost:   "{1}{U}",
		OracleText: `Other Wizards you control have "{T}: Draw a card, then discard a card."`,
	})
	effect := staticGrantContinuousEffect(t, face)
	subtypes := effect.Group.Selection().SubtypesAny
	if len(subtypes) != 1 || subtypes[0] != types.Wizard {
		t.Fatalf("group subtypes = %#v, want [Wizard]", subtypes)
	}
	if _, excluded := effect.Group.Exclusion(); !excluded {
		t.Fatalf("group = %#v, want source exclusion for an \"other\" lord", effect.Group)
	}
	if len(effect.AddAbilities) != 1 {
		t.Fatalf("added abilities = %#v, want exactly one granted ability", effect.AddAbilities)
	}
	if _, ok := effect.AddAbilities[0].(*game.ActivatedAbility); !ok {
		t.Fatalf("granted ability = %#v, want an activated ability", effect.AddAbilities[0])
	}
}

// TestLowerControlledArtifactGroupQuotedAbilityGrant verifies that a non-mana
// quoted ability granted to "Artifacts you control" lowers to an ability-layer
// continuous effect whose group requires the artifact card type.
func TestLowerControlledArtifactGroupQuotedAbilityGrant(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Artifact Lord",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		ManaCost:   "{2}{W}",
		OracleText: `Artifacts you control have "{1}, {T}: Draw a card."`,
	})
	effect := staticGrantContinuousEffect(t, face)
	required := effect.Group.Selection().RequiredTypes
	if len(required) != 1 || required[0] != types.Artifact {
		t.Fatalf("group required types = %#v, want [Artifact]", required)
	}
	if len(effect.AddAbilities) != 1 {
		t.Fatalf("added abilities = %#v, want exactly one granted ability", effect.AddAbilities)
	}
}
