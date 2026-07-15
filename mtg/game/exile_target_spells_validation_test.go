package game

import "testing"

func stackSpellTargetSpecForValidation() TargetSpec {
	return TargetSpec{
		MinTargets: 0,
		MaxTargets: 99,
		Constraint: "any number of target spells",
		Allow:      TargetAllowStackObject,
		Predicate:  TargetPredicate{StackObjectKinds: []StackObjectKind{StackSpell}},
	}
}

// TestAllTargetStackObjectsReferenceKindAndValidation proves the whole-spec
// group reference reports its kind and validates a non-negative spec index while
// rejecting a negative one.
func TestAllTargetStackObjectsReferenceKindAndValidation(t *testing.T) {
	t.Parallel()
	ref := AllTargetStackObjectsReference(0)
	if ref.Kind() != ObjectReferenceAllTargetStackObjects {
		t.Fatalf("kind = %v, want ObjectReferenceAllTargetStackObjects", ref.Kind())
	}
	specs := []TargetSpec{stackSpellTargetSpecForValidation()}
	if err := validateObjectReference(ref, specs, true); err != nil {
		t.Fatalf("valid all-target-stack-objects reference rejected: %v", err)
	}
	if err := firstProblem(AllTargetStackObjectsReference(-1).Validate()); err == nil {
		t.Fatal("negative spec index accepted")
	}
}

// TestExileTargetSpellsValidatesStackObjectReferences proves the primitive
// validates both the variable-count group reference and a single stack-object
// target reference against a stack-spell target spec.
func TestExileTargetSpellsValidatesStackObjectReferences(t *testing.T) {
	t.Parallel()
	specs := []TargetSpec{stackSpellTargetSpecForValidation()}
	if err := (ExileTargetSpells{Object: AllTargetStackObjectsReference(0)}).validatePrimitive(specs, true); err != nil {
		t.Fatalf("group exile validation failed: %v", err)
	}
	if err := (ExileTargetSpells{Object: TargetStackObjectReference(0)}).validatePrimitive(specs, true); err != nil {
		t.Fatalf("single-target exile validation failed: %v", err)
	}
}

// TestExileTargetSpellsRejectsNonStackReference proves the primitive fails closed
// when pointed at a non-stack-object reference: exile-target-spells only exiles
// spells on the stack, so a permanent target reference is rejected.
func TestExileTargetSpellsRejectsNonStackReference(t *testing.T) {
	t.Parallel()
	specs := []TargetSpec{stackSpellTargetSpecForValidation()}
	if err := (ExileTargetSpells{Object: TargetPermanentReference(0)}).validatePrimitive(specs, true); err == nil {
		t.Fatal("permanent target reference accepted by exile-target-spells")
	}
}

// TestExileTargetSpellsRejectsIncompatibleTargetSpec proves the group reference
// fails closed when its spec does not target stack objects: a permanent target
// spec cannot back an exile-target-spells group.
func TestExileTargetSpellsRejectsIncompatibleTargetSpec(t *testing.T) {
	t.Parallel()
	permanentSpec := TargetSpec{
		MinTargets: 0,
		MaxTargets: 99,
		Constraint: "any number of target creatures",
		Allow:      TargetAllowPermanent,
	}
	specs := []TargetSpec{permanentSpec}
	if err := (ExileTargetSpells{Object: AllTargetStackObjectsReference(0)}).validatePrimitive(specs, true); err == nil {
		t.Fatal("permanent target spec accepted by exile-target-spells group reference")
	}
}
