package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerTheSkullsporeNexus(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "The Skullspore Nexus",
		Layout:   "normal",
		ManaCost: "{6}{G}{G}",
		TypeLine: "Legendary Artifact",
		Colors:   []string{"G"},
		OracleText: "This spell costs {X} less to cast, where X is the greatest power among creatures you control.\n" +
			"Whenever one or more nontoken creatures you control die, create a green Fungus Dinosaur creature token with base power and toughness each equal to the total power of those creatures.\n" +
			"{2}, {T}: Double target creature's power until end of turn.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	create, ok := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("trigger effect = %#v, want CreateToken", face.TriggeredAbilities[0].Content)
	}
	if create.Amount.Value() != 1 || !create.Power.Exists || !create.Toughness.Exists {
		t.Fatalf("create token = %#v, want one dynamically sized token", create)
	}
	for name, size := range map[string]game.Quantity{"power": create.Power.Val, "toughness": create.Toughness.Val} {
		dynamic := size.DynamicAmount()
		if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountTriggeringEventTotalPower {
			t.Fatalf("%s = %#v, want triggering-event total power", name, size)
		}
	}
	def, ok := create.Source.TokenDefRef()
	if !ok ||
		def.Name != "Fungus Dinosaur" ||
		!slices.Equal(def.Colors, []color.Color{color.Green}) ||
		!slices.Equal(def.Types, []types.Card{types.Creature}) ||
		!slices.Equal(def.Subtypes, []types.Sub{types.Fungus, types.Dinosaur}) ||
		def.Power.Exists ||
		def.Toughness.Exists {
		t.Fatalf("token definition = %#v", def)
	}
}
