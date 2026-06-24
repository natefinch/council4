package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
)

// TestLowerMarwynSingleColorDynamicMana verifies that Marwyn, the Nurturer's
// "{T}: Add an amount of {G} equal to Marwyn's power." lowers to a mana ability
// that adds a fixed green mana in an amount equal to the source creature's
// power, the fixed-color sibling of the any-one-color dynamic mana path.
func TestLowerMarwynSingleColorDynamicMana(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Marwyn, the Nurturer",
		Layout:   "normal",
		TypeLine: "Legendary Creature — Elf Druid",
		ManaCost: "{3}{G}",
		OracleText: "Whenever another Elf you control enters, put a +1/+1 counter on Marwyn.\n" +
			"{T}: Add an amount of {G} equal to Marwyn's power.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("mana abilities = %d, want 1", len(face.ManaAbilities))
	}
	sequence := face.ManaAbilities[0].Content.Modes[0].Sequence
	if len(sequence) != 1 {
		t.Fatalf("sequence = %#v, want a single AddMana", sequence)
	}
	add, ok := sequence[0].Primitive.(game.AddMana)
	if !ok || add.ManaColor != mana.G || !add.Amount.IsDynamic() {
		t.Fatalf("primitive = %#v, want fixed green dynamic AddMana", sequence[0].Primitive)
	}
	dynamic := add.Amount.DynamicAmount().Val
	if dynamic.Kind != game.DynamicAmountObjectPower {
		t.Fatalf("dynamic amount kind = %v, want object power", dynamic.Kind)
	}
	if len(dynamic.Object.Validate()) != 0 {
		t.Fatalf("dynamic amount object invalid: %#v", dynamic.Object)
	}
	if err := game.ValidateInstructionSequence(sequence); err != nil {
		t.Fatalf("instruction sequence invalid: %v", err)
	}
}

// TestLowerSingleColorDynamicManaGreatestPower verifies the fixed-color dynamic
// mana path generalizes over the dynamic amount rather than binding to source
// power alone: Bighorner Rancher taps for green equal to the greatest power
// among creatures the controller controls.
func TestLowerSingleColorDynamicManaGreatestPower(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Bighorner Rancher",
		Layout:   "normal",
		TypeLine: "Creature — Human Ranger",
		ManaCost: "{4}{G}",
		OracleText: "Vigilance\n" +
			"{T}: Add an amount of {G} equal to the greatest power among creatures you control.",
	})
	if len(face.ManaAbilities) != 1 {
		t.Fatalf("mana abilities = %d, want 1", len(face.ManaAbilities))
	}
	sequence := face.ManaAbilities[0].Content.Modes[0].Sequence
	add, ok := sequence[len(sequence)-1].Primitive.(game.AddMana)
	if !ok || add.ManaColor != mana.G || !add.Amount.IsDynamic() {
		t.Fatalf("primitive = %#v, want fixed green dynamic AddMana", sequence[len(sequence)-1].Primitive)
	}
	if dynamic := add.Amount.DynamicAmount().Val; dynamic.Kind != game.DynamicAmountGreatestPowerInGroup {
		t.Fatalf("dynamic amount kind = %v, want greatest power in group", dynamic.Kind)
	}
}
