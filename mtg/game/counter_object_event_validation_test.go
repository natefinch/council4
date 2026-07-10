package game

import "testing"

func TestCounterObjectAllowsEventStackObjectReference(t *testing.T) {
	t.Parallel()
	ref := EventStackObjectReference()
	if ref.Kind() != ObjectReferenceEventStackObject {
		t.Fatalf("reference kind = %v, want event stack object", ref.Kind())
	}
	if err := validateObjectReference(ref, nil, true); err != nil {
		t.Fatalf("event stack object reference validation failed: %v", err)
	}
	if err := (CounterObject{Object: ref}).validatePrimitive(nil, true); err != nil {
		t.Fatalf("counter object validation failed: %v", err)
	}
}
