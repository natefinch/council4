package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/compare"
)

func TestNumericConditionLoweringUsesTypedPayload(t *testing.T) {
	t.Parallel()
	c := compiler.CompiledCondition{
		Kind: compiler.ConditionIf, Resolving: true,
		Predicate:           compiler.ConditionPredicateObjectMatches,
		HasSubjectReference: true, SubjectRefID: 17, SubjectPast: true, SubjectSpell: true,
		ObjectBinding:   compiler.ReferenceBindingTarget,
		ObjectReference: &compiler.CompiledReference{NodeID: 17, Binding: compiler.ReferenceBindingTarget, Occurrence: 1},
		ObjectTarget: &compiler.CompiledTarget{
			Cardinality: compiler.TargetCardinality{Min: 1, Max: 1},
			Selector:    compiler.CompiledSelector{Kind: compiler.SelectorSpell},
		},
		Selection: compiler.ConditionSelection{AttributeCompare: compiler.ConditionAttributeComparison{
			Attribute: compiler.ConditionAttributeManaValue, Op: compare.LessOrEqual, Value: 3,
		}},
	}
	expected, ok := lowerCondition(c, conditionContextEffectGate)
	if !ok {
		t.Fatal("typed comparison did not lower")
	}
	c.Text = "not Oracle"
	c.Span = shared.Span{Start: shared.Position{Offset: 900}, End: shared.Position{Offset: 1}}
	c.SubjectSpan = c.Span
	c.ObjectReference.Text = "not a spell"
	c.ObjectReference.Span = c.Span
	c.ObjectTarget.Text = "not a target"
	c.ObjectTarget.Span = c.Span
	got, ok := lowerCondition(c, conditionContextEffectGate)
	got.Text, expected.Text = "", ""
	if !ok || !reflect.DeepEqual(got, expected) {
		t.Fatalf("retained text/positions changed numeric semantics: %#v", got)
	}
	c.ObjectReference = nil
	if _, ok := lowerCondition(c, conditionContextEffectGate); ok {
		t.Fatal("missing exact reference payload lowered")
	}
}
