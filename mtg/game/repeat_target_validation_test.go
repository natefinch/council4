package game

import "testing"

func TestRepeatProcessInheritsEnclosingTargetValidation(t *testing.T) {
	t.Parallel()
	body := Mode{Sequence: []Instruction{{
		Primitive: Explore{Creature: TargetPermanentReference(0)},
	}}}.Ability()
	repeat := RepeatProcess{Times: Fixed(2), Body: body}
	specs := []TargetSpec{{Allow: TargetAllowPermanent, MinTargets: 1, MaxTargets: 1}}
	if err := repeat.validatePrimitive(specs, true); err != nil {
		t.Fatal(err)
	}
	if len(body.Modes[0].Targets) != 0 {
		t.Fatal("validation mutated the stored body")
	}
	if err := repeat.validatePrimitive(nil, true); err == nil {
		t.Fatal("missing enclosing target specification accepted")
	}
	badBody := Mode{Sequence: []Instruction{{
		Primitive: Explore{Creature: TargetPermanentReference(1)},
	}}}.Ability()
	if err := (RepeatProcess{Times: Fixed(2), Body: badBody}).validatePrimitive(specs, true); err == nil {
		t.Fatal("out-of-bounds body reference accepted")
	}
}
