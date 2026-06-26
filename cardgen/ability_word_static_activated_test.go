package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
)

// TestLowerCohortAbilityWord verifies that the Cohort ability word lowers as a
// rules-free label on an activated ability. Cohort adds no rules of its own (CR
// 207.2c uses the word purely for flavor on Ally cards), so the activated
// ability is the entire body: tapping the source and an untapped Ally you
// control are the activation costs and the effect taps a target creature.
func TestLowerCohortAbilityWord(t *testing.T) {
	t.Parallel()
	power := "2"
	toughness := "4"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Cohort Mage",
		Layout:     "normal",
		TypeLine:   "Creature — Human Wizard Ally",
		ManaCost:   "{3}{W}",
		OracleText: "Cohort — {T}, Tap an untapped Ally you control: Tap target creature.",
		Power:      &power,
		Toughness:  &toughness,
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if len(ability.AdditionalCosts) != 2 {
		t.Fatalf("additional costs = %d, want 2", len(ability.AdditionalCosts))
	}
	if ability.AdditionalCosts[0].Kind != cost.AdditionalTap {
		t.Fatalf("first additional cost = %v, want AdditionalTap", ability.AdditionalCosts[0].Kind)
	}
	if ability.AdditionalCosts[1].Kind != cost.AdditionalTapPermanents {
		t.Fatalf("second additional cost = %v, want AdditionalTapPermanents", ability.AdditionalCosts[1].Kind)
	}
}

// TestLowerFatefulHourAbilityWord verifies that the Fateful hour ability word
// lowers as a rules-free label on a conditional static ability. Fateful hour is
// flavor only; the rules live in the "as long as you have 5 or less life"
// condition gating the continuous power/toughness modification.
func TestLowerFatefulHourAbilityWord(t *testing.T) {
	t.Parallel()
	power := "1"
	toughness := "4"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Fateful Soldier",
		Layout:     "normal",
		TypeLine:   "Creature — Human Soldier",
		ManaCost:   "{2}{W}",
		OracleText: "Fateful hour — As long as you have 5 or less life, other creatures you control get +1/+4.",
		Power:      &power,
		Toughness:  &toughness,
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	static := face.StaticAbilities[0].Body
	if !static.Condition.Exists {
		t.Fatal("expected a gating condition on the Fateful hour static ability")
	}
	aggregates := static.Condition.Val.Aggregates
	if len(aggregates) != 1 ||
		aggregates[0].Op != compare.LessOrEqual ||
		aggregates[0].Value != 5 {
		t.Fatalf("condition aggregates = %#v, want controller life <= 5", aggregates)
	}
}

// TestLowerUndergrowthAbilityWord verifies that the Undergrowth ability word
// lowers as a rules-free label on a static spell cost-reduction ability. The
// word is flavor; the cost reduction per creature card in the graveyard is the
// only rule.
func TestLowerUndergrowthAbilityWord(t *testing.T) {
	t.Parallel()
	power := "6"
	toughness := "6"
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Undergrowth Hulk",
		Layout:     "normal",
		TypeLine:   "Creature — Plant Fungus Zombie",
		ManaCost:   "{6}{B}{G}",
		OracleText: "Undergrowth — This spell costs {1} less to cast for each creature card in your graveyard.",
		Power:      &power,
		Toughness:  &toughness,
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	rules := face.StaticAbilities[0].Body.RuleEffects
	if len(rules) != 1 || rules[0].Kind != game.RuleEffectCostModifier {
		t.Fatalf("rule effects = %#v, want one cost modifier", rules)
	}
}

// TestLowerUnknownStaticAbilityWordFailsClosed guards the rules-free gate: a
// flavor label that is not in the curated whitelist still fails closed on a
// static ability so only known rules-free ability words lower.
func TestLowerUnknownStaticAbilityWordFailsClosed(t *testing.T) {
	t.Parallel()
	power := "1"
	toughness := "4"
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Unknown Static Word",
		Layout:     "normal",
		TypeLine:   "Creature — Human Soldier",
		ManaCost:   "{2}{W}",
		OracleText: "Mysteryword — As long as you have 5 or less life, other creatures you control get +1/+4.",
		Power:      &power,
		Toughness:  &toughness,
	})
}
