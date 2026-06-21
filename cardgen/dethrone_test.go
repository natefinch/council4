package cardgen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerDethroneKeywordExpandsToTriggeredAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Treasonous Ogre",
		Layout:     "normal",
		TypeLine:   "Creature — Human Shaman",
		OracleText: "Dethrone (Whenever this creature attacks the player with the most life or tied for most life, put a +1/+1 counter on this creature.)",
	})
	if len(face.TriggeredAbilities) != 1 || len(face.StaticAbilities) != 0 {
		t.Fatalf("abilities = triggered:%d static:%d; want one triggered ability", len(face.TriggeredAbilities), len(face.StaticAbilities))
	}
	if !reflect.DeepEqual(face.TriggeredAbilities[0], game.DethroneTriggeredBody) {
		t.Fatalf("triggered ability = %+v; want game.DethroneTriggeredBody", face.TriggeredAbilities[0])
	}
}

func TestGenerateExecutableDethroneSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Scourge of the Throne",
		Layout:     "normal",
		TypeLine:   "Creature — Dragon",
		OracleText: "Flying\nDethrone (Whenever this creature attacks the player with the most life or tied for most life, put a +1/+1 counter on this creature.)",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"TriggeredAbilities: []game.TriggeredAbility",
		"game.DethroneTriggeredBody",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
