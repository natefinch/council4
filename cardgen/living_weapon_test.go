package cardgen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerLivingWeaponKeywordExpandsToTriggeredAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Flayer Husk",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Living weapon (When this Equipment enters, create a 0/0 black Phyrexian Germ creature token, then attach this Equipment to it.)\nEquipped creature gets +1/+1.\nEquip {1}",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d; want 1", len(face.TriggeredAbilities))
	}
	if !reflect.DeepEqual(face.TriggeredAbilities[0], game.LivingWeaponTriggeredAbility()) {
		t.Fatalf("triggered ability = %+v; want game.LivingWeaponTriggeredAbility()", face.TriggeredAbilities[0])
	}
}

func TestGenerateExecutableLivingWeaponSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Flayer Husk",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Living weapon (When this Equipment enters, create a 0/0 black Phyrexian Germ creature token, then attach this Equipment to it.)\nEquipped creature gets +1/+1.\nEquip {1}",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"TriggeredAbilities: []game.TriggeredAbility",
		"game.LivingWeaponTriggeredAbility()",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
