package game

import "testing"

func TestChangeStackObjectControllerValidation(t *testing.T) {
	t.Parallel()
	specs := []TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      TargetAllowStackObject,
		Predicate:  TargetPredicate{StackObjectKinds: []StackObjectKind{StackSpell}},
	}}
	primitive := ChangeStackObjectController{
		Object:     TargetStackObjectReference(0),
		Controller: ControllerReference(),
	}
	if err := primitive.validatePrimitive(specs, true); err != nil {
		t.Fatalf("valid primitive rejected: %v", err)
	}
	primitive.Object = TargetPermanentReference(0)
	if err := primitive.validatePrimitive(specs, true); err == nil {
		t.Fatal("permanent reference accepted")
	}
}
