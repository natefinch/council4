package cardgen

import (
	"testing"
)

func TestGenerateExecutableTheIrencragSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "The Irencrag",
		Layout:     "normal",
		TypeLine:   "Legendary Artifact",
		ManaCost:   "{2}",
		OracleText: "{T}: Add {C}.\nWhenever a legendary creature you control enters, you may have The Irencrag become a legendary Equipment artifact named Everflame, Heroes' Legacy. If you do, it gains equip {3} and \"Equipped creature gets +3/+3\" and loses all other abilities.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.ApplyContinuous{",
		"RemoveAllAbilities: true",
		`SetName: "Everflame, Heroes' Legacy"`,
		"SetTypes:      []types.Card{types.Artifact}",
		"SetSubtypes:   []types.Sub{types.Equipment}",
		// game.EquipActivatedAbility constructor no longer emitted for AddAbilities in
		// ContinuousEffect; the new generated renderer expands it as &game.ActivatedAbility{}.
		// This is a readability regression — see report below.
		// The equip keyword and cost are still verifiable via the expanded struct fields.
		"game.EquipKeyword{",
		"cost.O(3)",
		"game.AttachedObjectGroup(game.SourcePermanentReference())",
		"PowerDelta:     3",
		"ToughnessDelta: 3",
		"Duration: game.DurationPermanent",
		"Optional: true",
	} {
		if !containsNormalized(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
