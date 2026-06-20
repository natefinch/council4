package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerCumulativeUpkeepExactMana(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Remora",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Cumulative upkeep {1}{U}",
	})
	if len(face.TriggeredAbilities) != 1 || len(face.StaticAbilities) != 0 {
		t.Fatalf("abilities = triggered:%d static:%d; want one triggered ability", len(face.TriggeredAbilities), len(face.StaticAbilities))
	}
	ability := face.TriggeredAbilities[0]
	keyword, ok := game.BodyKeywordAbility(&ability, game.CumulativeUpkeep)
	if !ok {
		t.Fatal("lowered ability has no cumulative upkeep keyword")
	}
	cumulative, ok := keyword.(game.CumulativeUpkeepKeyword)
	if !ok || !slices.Equal(cumulative.Cost, cost.Mana{cost.O(1), cost.U}) {
		t.Fatalf("keyword = %+v; want exact {1}{U}", keyword)
	}
}

func TestLowerCumulativeUpkeepUnsupportedCostsFailClosed(t *testing.T) {
	t.Parallel()
	for _, oracleText := range []string{
		"Cumulative upkeep",
		"Cumulative upkeep {X}",
		"Cumulative upkeep—Pay 1 life.",
		"Cumulative upkeep {1} and pay 1 life.",
		"Cumulative upkeep {1} or {U}",
		"Cumulative upkeep cost {1}",
	} {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:       "Unsupported Upkeep",
				Layout:     "normal",
				TypeLine:   "Enchantment",
				OracleText: oracleText,
			})
			if len(face.TriggeredAbilities) != 0 || len(face.StaticAbilities) != 0 {
				t.Fatalf("unsupported %q partially lowered: triggered=%+v static=%+v", oracleText, face.TriggeredAbilities, face.StaticAbilities)
			}
		})
	}
}

func TestLowerMysticRemoraReusesGenericSpellTaxSupport(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Mystic Remora",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Cumulative upkeep {1}\nWhenever an opponent casts a noncreature spell, you may draw a card unless that player pays {4}.",
	})
	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("triggered abilities = %+v; want cumulative upkeep and spell tax", face.TriggeredAbilities)
	}
	if !game.BodyHasKeyword(&face.TriggeredAbilities[0], game.CumulativeUpkeep) {
		t.Fatalf("first trigger = %+v; want cumulative upkeep", face.TriggeredAbilities[0])
	}
	tax := face.TriggeredAbilities[1]
	if tax.Trigger.Pattern.Event != game.EventSpellCast ||
		tax.Trigger.Pattern.Controller != game.TriggerControllerOpponent ||
		!slices.Equal(tax.Trigger.Pattern.CardSelection.ExcludedTypes, []types.Card{types.Creature}) {
		t.Fatalf("tax trigger pattern = %+v; want opponent noncreature spell", tax.Trigger.Pattern)
	}
	sequence := game.BodyContent(&tax).Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("tax sequence = %+v; want payment and draw", sequence)
	}
	pay, ok := sequence[0].Primitive.(game.Pay)
	if !ok || !pay.Payment.ManaCost.Exists || !slices.Equal(pay.Payment.ManaCost.Val, cost.Mana{cost.O(4)}) {
		t.Fatalf("tax payment = %+v; want {4}", sequence[0])
	}
}

func TestGenerateExecutableCumulativeUpkeepSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Remora",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Cumulative upkeep {1}{U}",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"TriggeredAbilities: []game.TriggeredAbility",
		"game.CumulativeUpkeepTriggeredAbility(cost.Mana{",
		"cost.O(1)",
		"cost.U",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
