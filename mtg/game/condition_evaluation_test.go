package game

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/opt"
)

func TestValidateConditionEvaluations(t *testing.T) {
	t.Parallel()
	publisher := Instruction{
		Primitive:        Draw{Amount: Fixed(1), Player: ControllerReference()},
		Condition:        opt.Val(EffectCondition{Condition: opt.Val(Condition{ControllerHandEmpty: true})}),
		PublishCondition: "group",
	}
	consumer := Instruction{
		Primitive: GainLife{Amount: Fixed(1), Player: ControllerReference()}, ConditionGate: "group",
	}
	for _, tt := range []struct {
		name     string
		sequence []Instruction
		want     string
	}{
		{"valid", []Instruction{publisher, consumer}, ""},
		{"forward", []Instruction{consumer, publisher}, "not yet published"},
		{"missing", []Instruction{consumer}, "not yet published"},
		{"negation without gate", []Instruction{{Primitive: consumer.Primitive, ConditionGateNegate: true}}, "requires ConditionGate"},
		{"duplicate", []Instruction{publisher, publisher}, "duplicate condition key"},
		{"without predicate", []Instruction{{Primitive: publisher.Primitive, PublishCondition: "group"}}, "requires Condition"},
		{"nested primitive-only body", []Instruction{{
			Primitive: publisher.Primitive, TemptingOfferBody: []Instruction{publisher},
		}}, "inside TemptingOfferBody"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateInstructionSequence(tt.sequence)
			if tt.want == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error=%v, want %q", err, tt.want)
			}
		})
	}
}
