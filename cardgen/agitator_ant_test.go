package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

func agitatorAntCard() *ScryfallCard {
	power, toughness := "2", "2"
	return &ScryfallCard{
		Name:       "Agitator Ant",
		Layout:     "normal",
		ManaCost:   "{2}{R}",
		TypeLine:   "Creature — Insect",
		OracleText: "At the beginning of your end step, each player may put two +1/+1 counters on a creature they control. Goad each creature that had counters put on it this way. (Until your next turn, those creatures attack each combat if able and attack a player other than you if able.)",
		Power:      &power,
		Toughness:  &toughness,
	}
}

func TestLowerAgitatorAntOptionalCountersAndGoad(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, agitatorAntCard())
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want counter then goad", mode.Sequence)
	}
	place, ok := mode.Sequence[0].Primitive.(game.OptionalCounterForEachPlayer)
	if !ok {
		t.Fatalf("primitive[0] = %T", mode.Sequence[0].Primitive)
	}
	if place.Players != game.AllPlayersReference() ||
		place.Amount != game.Fixed(2) ||
		place.CounterKind != counter.PlusOnePlusOne ||
		place.PublishLinked == "" ||
		!slices.Contains(place.Selection.RequiredTypes, types.Creature) {
		t.Fatalf("placement = %#v", place)
	}
	goad, ok := mode.Sequence[1].Primitive.(game.Goad)
	if !ok || !goad.ConsumeLinked {
		t.Fatalf("primitive[1] = %#v, want consuming goad", mode.Sequence[1].Primitive)
	}
	if key, linked := goad.Group.LinkedKey(); !linked || key != place.PublishLinked {
		t.Fatalf("goad group key = (%q, %t), placement key = %q", key, linked, place.PublishLinked)
	}
}

func TestGenerateExecutableAgitatorAnt(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(agitatorAntCard(), "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.OptionalCounterForEachPlayer{",
		"game.AllPlayersReference()",
		"game.Fixed(2)",
		"counter.PlusOnePlusOne",
		"game.LinkedObjectsGroup(game.LinkedKey(\"optional-counter-for-each-player\"))",
		"ConsumeLinked: true",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
