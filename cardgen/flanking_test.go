package cardgen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerFlankingKeywordExpandsToTriggeredAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Benalish Cavalry",
		Layout:     "normal",
		TypeLine:   "Creature — Human Knight",
		OracleText: "Flanking (Whenever a creature without flanking blocks this creature, the blocking creature gets -1/-1 until end of turn.)",
	})
	if len(face.TriggeredAbilities) != 1 || len(face.StaticAbilities) != 0 {
		t.Fatalf("abilities = triggered:%d static:%d; want one triggered ability", len(face.TriggeredAbilities), len(face.StaticAbilities))
	}
	if !reflect.DeepEqual(face.TriggeredAbilities[0], game.FlankingTriggeredBody) {
		t.Fatalf("triggered ability = %+v; want game.FlankingTriggeredBody", face.TriggeredAbilities[0])
	}
}

func TestGenerateExecutableFlankingSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Benalish Cavalry",
		Layout:     "normal",
		TypeLine:   "Creature — Human Knight",
		OracleText: "Flanking (Whenever a creature without flanking blocks this creature, the blocking creature gets -1/-1 until end of turn.)",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"TriggeredAbilities: []game.TriggeredAbility",
		"game.FlankingTriggeredBody",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
