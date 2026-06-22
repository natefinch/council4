package cardgen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerStartEnginesKeywordExpandsToTriggeredAbility verifies that the
// "Start your engines!" keyword lowers to the shared StartEnginesTriggeredBody.
func TestLowerStartEnginesKeywordExpandsToTriggeredAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Engine Starter",
		Layout:   "normal",
		TypeLine: "Creature — Human Pilot",
		OracleText: "Start your engines! (If you have no speed, it starts at 1. " +
			"It increases once on each of your turns when an opponent loses life. Max speed is 4.)",
	})
	if len(face.TriggeredAbilities) != 1 || len(face.StaticAbilities) != 0 {
		t.Fatalf("abilities = triggered:%d static:%d; want one triggered ability",
			len(face.TriggeredAbilities), len(face.StaticAbilities))
	}
	if !reflect.DeepEqual(face.TriggeredAbilities[0], game.StartEnginesTriggeredBody) {
		t.Fatalf("triggered ability = %+v; want game.StartEnginesTriggeredBody", face.TriggeredAbilities[0])
	}
}

// TestGenerateExecutableTheSpeedDemon confirms the anchor card fully generates:
// the "Start your engines!" keyword, plus an end-step ability whose draw and
// life-loss both read the controller's speed (the shared "where X is your
// speed" amount binds to both clauses).
func TestGenerateExecutableTheSpeedDemon(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:          "The Speed Demon",
		Layout:        "normal",
		ManaCost:      "{3}{B}{B}",
		TypeLine:      "Legendary Creature — Demon",
		ColorIdentity: []string{"B"},
		Power:         new("5"),
		Toughness:     new("5"),
		OracleText: "Flying, trample\nStart your engines! (If you have no speed, it starts at 1. " +
			"It increases once on each of your turns when an opponent loses life. Max speed is 4.)\n" +
			"At the beginning of your end step, you draw X cards and lose X life, where X is your speed.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.FlyingStaticBody",
		"game.TrampleStaticBody",
		"game.StartEnginesTriggeredBody",
		"game.Draw{",
		"game.LoseLife{",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
	// Both the draw and the life loss must read the controller's speed: a single
	// DynamicAmountControllerSpeed appears once per clause (exactly twice).
	if got := strings.Count(source, "game.DynamicAmountControllerSpeed"); got != 2 {
		t.Fatalf("DynamicAmountControllerSpeed count = %d, want 2 (draw and lose):\n%s", got, source)
	}
}
