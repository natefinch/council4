package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
)

// manaAbilityDamageRider returns the self-damage Damage instruction carried by a
// lowered mana ability, asserting it is the source-dealt, controller-targeting
// rider that painlands and similar mana sources print.
func manaAbilityDamageRider(t *testing.T, ability *game.ManaAbility) game.Damage {
	t.Helper()
	if len(ability.Content.Modes) != 1 {
		t.Fatalf("mana ability content modes = %d, want 1", len(ability.Content.Modes))
	}
	sequence := ability.Content.Modes[0].Sequence
	if len(sequence) == 0 {
		t.Fatal("mana ability content has no instructions")
	}
	damage, ok := sequence[len(sequence)-1].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("last instruction = %T, want game.Damage", sequence[len(sequence)-1].Primitive)
	}
	player, ok := damage.Recipient.PlayerReference()
	if !ok || player.Kind() != game.PlayerReferenceController {
		t.Fatalf("damage recipient = %#v, want controller player reference", damage.Recipient)
	}
	if !damage.DamageSource.Exists || damage.DamageSource.Val != game.SourcePermanentReference() {
		t.Fatalf("damage source = %#v, want source permanent", damage.DamageSource)
	}
	return damage
}

// expectUnsupportedManaRider asserts that a mana ability whose body carries a
// non-rider second effect fails closed with the mana-effect diagnostic and
// lowers no mana ability.
func expectUnsupportedManaRider(t *testing.T, oracleText string) {
	t.Helper()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Mana Source",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: oracleText,
	})
	for i := range faces {
		if len(faces[i].ManaAbilities) != 0 {
			t.Fatalf("%q unexpectedly lowered a mana ability", oracleText)
		}
	}
	found := false
	for i := range diagnostics {
		if diagnostics[i].Summary == "unsupported mana effect" {
			found = true
		}
	}
	if !found {
		t.Fatalf("diagnostics = %#v, want unsupported mana effect", diagnostics)
	}
}

// TestLowerPainlandSelfDamageRider verifies that the two-ability painland shape
// lowers both the colorless ability and the colored ability carrying the
// "deals 1 damage to you" rider.
func TestLowerPainlandSelfDamageRider(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Shivan Reef",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {C}.\n{T}: Add {U} or {R}. This land deals 1 damage to you.",
	})
	if len(face.ManaAbilities) != 2 {
		t.Fatalf("mana abilities = %d, want 2", len(face.ManaAbilities))
	}
	damage := manaAbilityDamageRider(t, &face.ManaAbilities[1])
	if damage.Amount != game.Fixed(1) {
		t.Fatalf("rider amount = %#v, want fixed 1", damage.Amount)
	}
	// The colored ability also offers the U/R color choice ahead of the rider.
	sequence := face.ManaAbilities[1].Content.Modes[0].Sequence
	if len(sequence) != 3 || sequence[0].Primitive.Kind() != game.PrimitiveChoose {
		t.Fatalf("colored ability sequence = %#v, want choose+add+damage", sequence)
	}
}

// TestLowerSingleColorSelfDamageRider verifies the Elves of Deep Shadow shape:
// a single mana ability adding one color and dealing one damage to the
// controller.
func TestLowerSingleColorSelfDamageRider(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Elves of Deep Shadow",
		Layout:     "normal",
		TypeLine:   "Creature — Elf Druid",
		OracleText: "{T}: Add {B}. This creature deals 1 damage to you.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("mana abilities = %d, want 1", len(face.ManaAbilities))
	}
	sequence := face.ManaAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence = %#v, want add+damage", sequence)
	}
	add, ok := sequence[0].Primitive.(game.AddMana)
	if !ok || add.ManaColor != mana.B {
		t.Fatalf("sequence[0] = %#v, want AddMana B", sequence[0].Primitive)
	}
	if damage := manaAbilityDamageRider(t, &face.ManaAbilities[0]); damage.Amount != game.Fixed(1) {
		t.Fatalf("rider amount = %#v, want fixed 1", damage.Amount)
	}
}

// TestLowerColorlessSelfDamageRiderAmount verifies the Ancient Tomb shape: two
// colorless mana with a two-damage rider, confirming the rider preserves the
// printed amount rather than collapsing to one.
func TestLowerColorlessSelfDamageRiderAmount(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Ancient Tomb",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {C}{C}. This land deals 2 damage to you.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("mana abilities = %d, want 1", len(face.ManaAbilities))
	}
	if damage := manaAbilityDamageRider(t, &face.ManaAbilities[0]); damage.Amount != game.Fixed(2) {
		t.Fatalf("rider amount = %#v, want fixed 2", damage.Amount)
	}
}

// TestLowerManaRiderFailsClosed confirms that only the self-damage-to-you rider
// is accepted; every other trailing mana-ability effect fails closed.
func TestLowerManaRiderFailsClosed(t *testing.T) {
	t.Parallel()
	cases := []string{
		"{T}: Add {G}. Draw a card.",
		"{T}: Add {R}. This land deals 1 damage to each opponent.",
		"{T}: Add {W}. You gain 1 life.",
		"{T}: Add {B}. You lose 1 life.",
	}
	for _, oracleText := range cases {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			expectUnsupportedManaRider(t, oracleText)
		})
	}
}
