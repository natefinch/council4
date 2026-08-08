package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
)

// TestLowerCarnivorousCanopyManaValueComparePermanentNoun guards the
// target-attribute-compare condition family with a real card: "Destroy
// target artifact, enchantment, or creature with flying. If that permanent's
// mana value was 3 or less, proliferate." (Carnivorous Canopy). The
// named-possessive "that permanent's" binds the gate's object to the
// destroy's own target, and the gate carries a mana-value less-or-equal
// comparison rather than any type/subtype restriction.
func TestLowerCarnivorousCanopyManaValueComparePermanentNoun(t *testing.T) {
	t.Parallel()
	sequence := lowerSpellSequence(t, "Carnivorous Canopy Test",
		"Destroy target artifact, enchantment, or creature with flying. If that permanent's mana value was 3 or less, proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)")
	if len(sequence) != 2 {
		t.Fatalf("sequence = %#v, want two instructions (destroy, gated proliferate)", sequence)
	}
	destroy := sequence[0]
	if _, ok := destroy.Primitive.(game.Destroy); !ok {
		t.Fatalf("instruction[0] = %T, want game.Destroy", destroy.Primitive)
	}
	if destroy.Condition.Exists {
		t.Fatalf("destroy must carry no gate of its own: %#v", destroy)
	}
	proliferate := sequence[1]
	gate := effectConditionMatch(t, proliferate)
	if gate.Object.Val.Kind() != game.ObjectReferenceTargetPermanent || gate.Object.Val.TargetIndex() != 0 {
		t.Fatalf("gate object = %#v, want target permanent 0", gate.Object)
	}
	if got := gate.ObjectMatches.Val.ManaValue; !got.Exists || got.Val != (compare.Int{Op: compare.LessOrEqual, Value: 3}) {
		t.Fatalf("gate mana value = %#v, want <= 3", got)
	}
	if gate.ObjectMatches.Val.Power.Exists || gate.ObjectMatches.Val.Toughness.Exists {
		t.Fatalf("gate must carry no power/toughness bound: %#v", gate)
	}
}

// TestLowerTargetAttributeComparePowerToughnessPermanentNoun covers the power
// and toughness attributes of the same family via the named-possessive form.
// Real cards use this grammar (Depressurize: "Target creature gets -3/-0
// until end of turn. Then if that creature's power is 0 or less, destroy
// it."; Gore Vassal: "... Then if that creature's toughness is 1 or greater,
// regenerate it."), but both currently hit an unrelated blocker in the
// gated consequence (an unsupported "destroy it"/"regenerate it" back-
// reference, not the condition itself), so this test isolates the condition
// gate alone with an otherwise-equivalent, fully supported gated effect.
func TestLowerTargetAttributeComparePowerToughnessPermanentNoun(t *testing.T) {
	t.Parallel()
	sequence := lowerSpellSequence(t, "Power Toughness Compare Test",
		"Destroy target creature. If that creature's power is 4 or greater, draw a card. If that creature's toughness is 1 or greater, you gain 1 life.")
	if len(sequence) != 3 {
		t.Fatalf("sequence = %#v, want three instructions (destroy, gated draw, gated life gain)", sequence)
	}
	draw := sequence[1]
	drawGate := effectConditionMatch(t, draw)
	if got := drawGate.ObjectMatches.Val.Power; !got.Exists || got.Val != (compare.Int{Op: compare.GreaterOrEqual, Value: 4}) {
		t.Fatalf("draw gate power = %#v, want >= 4", got)
	}
	if drawGate.ObjectMatches.Val.ManaValue.Exists || drawGate.ObjectMatches.Val.Toughness.Exists {
		t.Fatalf("draw gate must carry no mana-value/toughness bound: %#v", drawGate)
	}
	gainLife := sequence[2]
	lifeGate := effectConditionMatch(t, gainLife)
	if got := lifeGate.ObjectMatches.Val.Toughness; !got.Exists || got.Val != (compare.Int{Op: compare.GreaterOrEqual, Value: 1}) {
		t.Fatalf("life-gain gate toughness = %#v, want >= 1", got)
	}
}

// TestLowerTargetAttributeCompareSpellNounFailsClosed guards
// recognizeTargetAttributeCompareCondition's deliberate exclusion of a "that
// spell's" possessive antecedent: Reject Imperfection ("Counter target
// spell. If that spell's mana value was 3 or less, proliferate.") is not yet
// supported, because lowerObjectReference's Target case always produces a
// TargetPermanentReference with no path to a TargetStackObjectReference --
// binding a spell target that way would silently resolve the wrong reference
// kind rather than fail closed. The recognizer rejects the "spell" noun
// outright so the ability fails closed as unsupported instead.
func TestLowerTargetAttributeCompareSpellNounFailsClosed(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Reject Imperfection Probe",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Counter target spell. If that spell's mana value was 3 or less, proliferate.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "p")
	if err != nil {
		t.Fatalf("GenerateExecutableCardSource error = %v", err)
	}
	if len(diagnostics) == 0 && source != "" {
		t.Fatal("spell-noun attribute-compare gate unexpectedly compiled without any diagnostic")
	}
}

// TestLowerBareItsAttributeCompareNotClaimedByTarget guards against the
// recognizer regression this family's development uncovered: a bare "its
// power is/was N or greater" always binds to the triggering event's
// permanent via the pre-existing recognizeEventSubjectPowerState, never to
// the clause's own target, because a real, already-shipped card depends on
// that binding for a trigger body with no target at all (Tribute to the
// World Tree: "Whenever a creature you control enters, draw a card if its
// power is 3 or greater. Otherwise, put two +1/+1 counters on it."). This
// guards that a spell-sequence body using the SAME bare "its power ... or
// greater" wording after a target effect does not silently get bound to the
// (irrelevant) event-permanent reference instead of failing closed: with no
// event permanent available in a plain spell body, it should compile as
// unsupported, not target the wrong object.
func TestLowerBareItsAttributeCompareNotClaimedByTarget(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Bare Its Power Probe",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Destroy target creature. If its power is 4 or greater, draw a card.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "p")
	if err != nil {
		t.Fatalf("GenerateExecutableCardSource error = %v", err)
	}
	if len(diagnostics) == 0 && source != "" {
		t.Fatal("bare-its attribute-compare gate unexpectedly compiled without any diagnostic")
	}
}

// TestLowerNoTargetAttributeCompareFailsClosed guards the compiler's
// target-reference-binding validation (bindConditionReferences in
// cardgen/oracle/compiler/condition.go), which is the actual safety net for
// a possessive antecedent that parses to ConditionObjectBindingTarget but
// names no real target: "Whenever a creature you control enters, scry 1. If
// that creature's mana value is 3 or greater, draw a card." uses an allowed
// permanent-type noun ("creature"), so the condition parses successfully as
// ConditionPredicateObjectMatches/ConditionObjectBindingTarget (confirmed
// directly against the parser: Parse() on this exact text produces one such
// clause) -- but the ability has no target anywhere, so
// bindConditionReferences correctly rejects the binding and the ability
// fails closed as unsupported rather than silently resolving against
// nothing. This is deliberately NOT the Zaffai/magecraft shape used
// elsewhere in this file (TestLowerTargetAttributeCompareSpellNounFailsClosed):
// that card's "that spell's" is rejected by the parser's noun allowlist
// before any binding is ever attempted (Parse() on Zaffai's text produces
// zero condition clauses), so it does not exercise this validation step at
// all, even though its own doc comment once claimed it did.
func TestLowerNoTargetAttributeCompareFailsClosed(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "No Target Attribute Compare Probe",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a creature you control enters, scry 1. If that creature's mana value is 3 or greater, draw a card.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "p")
	if err != nil {
		t.Fatalf("GenerateExecutableCardSource error = %v", err)
	}
	if len(diagnostics) == 0 && source != "" {
		t.Fatal("no-target attribute-compare gate unexpectedly compiled without any diagnostic")
	}
}

// effectConditionMatch extracts the nested game.Condition an instruction's
// per-effect gate carries (its Object reference and ObjectMatches selection),
// failing the test if the instruction is not gated by exactly one such
// condition.
func effectConditionMatch(t *testing.T, instr game.Instruction) game.Condition {
	t.Helper()
	if !instr.Condition.Exists || !instr.Condition.Val.Condition.Exists || !instr.Condition.Val.Condition.Val.ObjectMatches.Exists {
		t.Fatalf("instruction = %#v, want a per-effect ObjectMatches gate", instr)
	}
	return instr.Condition.Val.Condition.Val
}
