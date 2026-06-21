package cardgen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerUndyingKeywordExpandsToTriggeredAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Young Wolf",
		Layout:     "normal",
		TypeLine:   "Creature — Wolf",
		OracleText: "Undying (When this creature dies, if it had no +1/+1 counters on it, return it to the battlefield under its owner's control with a +1/+1 counter on it.)",
	})
	if len(face.TriggeredAbilities) != 1 || len(face.StaticAbilities) != 0 {
		t.Fatalf("abilities = triggered:%d static:%d; want one triggered ability", len(face.TriggeredAbilities), len(face.StaticAbilities))
	}
	if !reflect.DeepEqual(face.TriggeredAbilities[0], game.UndyingTriggeredBody) {
		t.Fatalf("triggered ability = %+v; want game.UndyingTriggeredBody", face.TriggeredAbilities[0])
	}
}

func TestLowerPersistKeywordExpandsAlongsideOtherText(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Kitchen Finks",
		Layout:     "normal",
		TypeLine:   "Creature — Ouphe",
		OracleText: "When this creature enters, you gain 2 life.\nPersist (When this creature dies, if it had no -1/-1 counters on it, return it to the battlefield under its owner's control with a -1/-1 counter on it.)",
	})
	var persist *game.TriggeredAbility
	for i := range face.TriggeredAbilities {
		if reflect.DeepEqual(face.TriggeredAbilities[i], game.PersistTriggeredBody) {
			persist = &face.TriggeredAbilities[i]
		}
	}
	if persist == nil {
		t.Fatalf("triggered abilities = %+v; want one to equal game.PersistTriggeredBody", face.TriggeredAbilities)
	}
}

func TestGenerateExecutableUndyingSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Young Wolf",
		Layout:     "normal",
		TypeLine:   "Creature — Wolf",
		OracleText: "Undying (When this creature dies, if it had no +1/+1 counters on it, return it to the battlefield under its owner's control with a +1/+1 counter on it.)",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"TriggeredAbilities: []game.TriggeredAbility",
		"game.UndyingTriggeredBody",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
