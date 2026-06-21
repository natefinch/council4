package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerFabricateExpandsToEntryTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Weaponcraft Enthusiast",
		Layout:     "normal",
		TypeLine:   "Creature — Aetherborn Artificer",
		OracleText: "Fabricate 2 (When this creature enters, put two +1/+1 counters on it or create two 1/1 colorless Servo artifact creature tokens.)",
	})
	if len(face.TriggeredAbilities) != 1 || len(face.StaticAbilities) != 0 {
		t.Fatalf("abilities = triggered:%d static:%d; want one triggered ability", len(face.TriggeredAbilities), len(face.StaticAbilities))
	}
	ability := face.TriggeredAbilities[0]
	keyword, ok := game.BodyKeywordAbility(&ability, game.Fabricate)
	if !ok {
		t.Fatal("lowered ability has no fabricate keyword")
	}
	fabricate, ok := keyword.(game.FabricateKeyword)
	if !ok || fabricate.Count != 2 {
		t.Fatalf("keyword = %+v; want fabricate count 2", keyword)
	}
	if !ability.Content.IsModal() || len(ability.Content.Modes) != 2 {
		t.Fatalf("content = %+v; want a two-mode modal choice", ability.Content)
	}
}

func TestGenerateExecutableFabricateSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Visionary Augmenter",
		Layout:     "normal",
		TypeLine:   "Artifact Creature — Construct",
		OracleText: "Fabricate 1 (When this creature enters, put a +1/+1 counter on it or create a 1/1 colorless Servo artifact creature token.)",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"TriggeredAbilities: []game.TriggeredAbility",
		"game.FabricateTriggeredAbility(1)",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
