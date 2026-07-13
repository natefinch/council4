package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
)

func TestLowerEchoExactMana(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Echoer",
		Layout:     "normal",
		TypeLine:   "Creature — Bird",
		ManaCost:   "{2}{U}",
		OracleText: "Flying\nEcho {2}{U}",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d; want one echo trigger", len(face.TriggeredAbilities))
	}
	ability := face.TriggeredAbilities[0]
	keyword, ok := game.BodyKeywordAbility(&ability, game.Echo)
	if !ok {
		t.Fatal("lowered ability has no echo keyword")
	}
	echo, ok := keyword.(game.EchoKeyword)
	if !ok || !slices.Equal(echo.Cost, cost.Mana{cost.O(2), cost.U}) {
		t.Fatalf("keyword = %+v; want exact {2}{U}", keyword)
	}
	if ability.Trigger.Pattern.Event != game.EventBeginningOfStep ||
		ability.Trigger.Pattern.Step != game.StepUpkeep ||
		ability.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("trigger pattern = %+v; want your upkeep", ability.Trigger.Pattern)
	}
	if !ability.Trigger.InterveningCondition.Exists ||
		!ability.Trigger.InterveningCondition.Val.SourceCameUnderControlSinceLastUpkeep {
		t.Fatalf("intervening condition = %+v; want came-under-control since last upkeep", ability.Trigger.InterveningCondition)
	}
}

func TestLowerEchoZeroCost(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Free Echoer",
		Layout:     "normal",
		TypeLine:   "Creature — Efreet",
		ManaCost:   "{3}{R}",
		OracleText: "Trample\nEcho {0}",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d; want one echo trigger", len(face.TriggeredAbilities))
	}
	keyword, ok := game.BodyKeywordAbility(&face.TriggeredAbilities[0], game.Echo)
	if !ok {
		t.Fatal("lowered ability has no echo keyword")
	}
	echo, ok := keyword.(game.EchoKeyword)
	if !ok || !slices.Equal(echo.Cost, cost.Mana{cost.O(0)}) {
		t.Fatalf("keyword = %+v; want {0}", keyword)
	}
}

func TestLowerEchoNonManaCostsFailClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Echo—Discard a card.",
		"Echo—Sacrifice two lands.",
		"Echo {X}",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:       "Unsupported Echoer",
				Layout:     "normal",
				TypeLine:   "Creature — Elemental",
				ManaCost:   "{4}{R}{R}",
				OracleText: oracleText,
			})
			if len(face.TriggeredAbilities) != 0 || len(face.StaticAbilities) != 0 {
				t.Fatalf("unsupported %q partially lowered: triggered=%+v static=%+v", oracleText, face.TriggeredAbilities, face.StaticAbilities)
			}
		})
	}
}

func TestEchoFlavorNameAbilityNotScannedAsKeyword(t *testing.T) {
	t.Parallel()
	// A flavored ability name that begins with the word "Echo" but has no mana
	// cost following it must not be misread as the Echo keyword.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Flavor Echoer",
		Layout:     "normal",
		TypeLine:   "Artifact",
		ManaCost:   "{5}",
		OracleText: "Echo of the First Murder — When this artifact enters, exile up to one target creature.",
	})
	for _, ability := range face.TriggeredAbilities {
		if game.BodyHasKeyword(&ability, game.Echo) {
			t.Fatalf("flavor-named ability misread as echo keyword: %+v", ability)
		}
	}
}

func TestGenerateExecutableEchoSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Herald of Serra",
		Layout:     "normal",
		TypeLine:   "Creature — Angel",
		ManaCost:   "{2}{W}{W}",
		OracleText: "Flying, vigilance\nEcho {2}{W}{W}",
	}, "h")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"TriggeredAbilities: []game.TriggeredAbility",
		"game.EchoTriggeredAbility(cost.Mana{",
		"cost.O(2)",
		"cost.W",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
