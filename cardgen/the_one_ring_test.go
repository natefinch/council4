package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
)

const theOneRingOracle = "Indestructible\n" +
	"When The One Ring enters, if you cast it, you gain protection from everything until your next turn.\n" +
	"At the beginning of your upkeep, you lose 1 life for each burden counter on The One Ring.\n" +
	"{T}: Put a burden counter on The One Ring, then draw a card for each burden counter on The One Ring."

func theOneRingCard() *ScryfallCard {
	return &ScryfallCard{
		Name:       "The One Ring",
		Layout:     "normal",
		TypeLine:   "Legendary Artifact",
		OracleText: theOneRingOracle,
	}
}

func TestLowerTheOneRing(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, theOneRingCard())
	if len(face.StaticAbilities) != 1 ||
		len(face.TriggeredAbilities) != 2 ||
		len(face.ActivatedAbilities) != 1 {
		t.Fatalf(
			"abilities = static:%d triggered:%d activated:%d, want 1/2/1",
			len(face.StaticAbilities),
			len(face.TriggeredAbilities),
			len(face.ActivatedAbilities),
		)
	}

	enter := face.TriggeredAbilities[0]
	if !enter.Trigger.InterveningIfEventPermanentWasCastByController ||
		enter.Trigger.InterveningIfEventPermanentWasCast {
		t.Fatalf("enter trigger = %#v, want cast by its controller", enter.Trigger)
	}
	if len(enter.Content.Modes) != 1 || len(enter.Content.Modes[0].Sequence) != 1 {
		t.Fatalf("enter content = %#v, want one instruction", enter.Content)
	}
	apply, ok := enter.Content.Modes[0].Sequence[0].Primitive.(game.ApplyRule)
	if !ok || apply.Duration != game.DurationUntilYourNextTurn ||
		len(apply.RuleEffects) != 1 ||
		apply.RuleEffects[0].Kind != game.RuleEffectPlayerProtection ||
		apply.RuleEffects[0].AffectedPlayer != game.PlayerYou ||
		!apply.RuleEffects[0].Protection.Everything {
		t.Fatalf("enter primitive = %#v, want player protection until next turn", enter.Content.Modes[0].Sequence[0].Primitive)
	}

	upkeep := face.TriggeredAbilities[1]
	if len(upkeep.Content.Modes) != 1 || len(upkeep.Content.Modes[0].Sequence) != 1 {
		t.Fatalf("upkeep content = %#v, want one instruction", upkeep.Content)
	}
	lose, ok := upkeep.Content.Modes[0].Sequence[0].Primitive.(game.LoseLife)
	if !ok {
		t.Fatalf("upkeep primitive = %T, want game.LoseLife", upkeep.Content.Modes[0].Sequence[0].Primitive)
	}
	assertBurdenCounterAmount(t, lose.Amount)

	activated := face.ActivatedAbilities[0]
	if len(activated.AdditionalCosts) != 1 || activated.AdditionalCosts[0].Kind != cost.AdditionalTap {
		t.Fatalf("activation costs = %#v, want tap", activated.AdditionalCosts)
	}
	sequence := activated.Content.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("activation sequence = %#v, want counter then draw", sequence)
	}
	add, ok := sequence[0].Primitive.(game.AddCounter)
	if !ok || add.CounterKind != counter.Burden || add.Object != game.SourcePermanentReference() {
		t.Fatalf("first primitive = %#v, want burden counter on source", sequence[0].Primitive)
	}
	draw, ok := sequence[1].Primitive.(game.Draw)
	if !ok {
		t.Fatalf("second primitive = %T, want draw", sequence[1].Primitive)
	}
	assertBurdenCounterAmount(t, draw.Amount)
}

func TestGenerateExecutableTheOneRingSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(theOneRingCard(), "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"InterveningIfEventPermanentWasCastByController: true",
		"game.RuleEffectPlayerProtection",
		"game.DurationUntilYourNextTurn",
		"counter.Burden",
		"game.DynamicAmountObjectCounters",
		"AdditionalCosts: cost.Tap",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestLowerPlayerProtectionFailsClosed(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		"When Test Relic enters, if you cast it, you gain protection from red until your next turn.",
		"When Test Relic enters, if you cast it, you gain protection from everything until end of turn.",
		"When Test Relic enters, if an opponent cast it, you gain protection from everything until your next turn.",
	} {
		t.Run(oracle, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:       "Test Relic",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: oracle,
			})
			if len(face.TriggeredAbilities) != 0 {
				t.Fatalf("unsupported effect lowered: %#v", face.TriggeredAbilities)
			}
		})
	}
}

func TestLowerSourceCounterAmountFailsClosed(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		"At the beginning of your upkeep, you lose 1 life for each mystery counter on Test Relic.",
		"At the beginning of your upkeep, you lose 1 life for each burden counter on target artifact.",
		"{T}: Put a burden counter on Test Relic, then draw a card for each burden counters on Test Relic.",
	} {
		t.Run(oracle, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:       "Test Relic",
				Layout:     "normal",
				TypeLine:   "Artifact",
				OracleText: oracle,
			})
			if len(face.TriggeredAbilities) != 0 || len(face.ActivatedAbilities) != 0 {
				t.Fatalf("unsupported amount lowered: %#v", face)
			}
		})
	}
}

func assertBurdenCounterAmount(t *testing.T, amount game.Quantity) {
	t.Helper()
	dynamic := amount.DynamicAmount()
	if !dynamic.Exists ||
		dynamic.Val.Kind != game.DynamicAmountObjectCounters ||
		dynamic.Val.Object != game.SourcePermanentReference() ||
		dynamic.Val.CounterKind != counter.Burden {
		t.Fatalf("amount = %#v, want source burden counter count", amount)
	}
}
