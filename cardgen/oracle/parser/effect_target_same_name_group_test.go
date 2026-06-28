package parser

import "testing"

// TestParseSameNameGroupTargetIsExact verifies that the trailing "and all other
// <type> with the same name as that <noun>" clause is captured onto the target's
// selection as a SameNameGroup rather than leaking into the effect or spoiling
// the exact, lowerable target ("Destroy target nonland permanent and all other
// permanents with the same name as that permanent" — Maelstrom Pulse).
func TestParseSameNameGroupTargetIsExact(t *testing.T) {
	t.Parallel()
	targets := firstDestroyTargets(t, "Destroy target nonland permanent and all other permanents with the same name as that permanent.")
	if len(targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(targets))
	}
	target := targets[0]
	if target.Selection.Kind != SelectionPermanent {
		t.Fatalf("target kind = %v, want SelectionPermanent", target.Selection.Kind)
	}
	if target.Selection.SameNameGroup == nil {
		t.Fatal("target SameNameGroup = nil, want non-nil")
	}
	if len(target.Selection.SameNameGroup.GroupTypes) != 0 {
		t.Fatalf("permanent group types = %v, want empty (no card-type restriction)", target.Selection.SameNameGroup.GroupTypes)
	}
	if !target.Exact {
		t.Fatal("target Exact = false, want true")
	}
}

// TestParseSameNameGroupTypedTarget verifies that a typed sibling of the cycle
// records the printed group card type ("all other lands ...") while still
// round-tripping to an exact target (Wake of Destruction, the Echoing cycle).
func TestParseSameNameGroupTypedTarget(t *testing.T) {
	t.Parallel()
	targets := firstDestroyTargets(t, "Destroy target land and all other lands with the same name as that land.")
	if len(targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(targets))
	}
	target := targets[0]
	if target.Selection.Kind != SelectionLand {
		t.Fatalf("target kind = %v, want SelectionLand", target.Selection.Kind)
	}
	group := target.Selection.SameNameGroup
	if group == nil {
		t.Fatal("target SameNameGroup = nil, want non-nil")
	}
	if len(group.GroupTypes) != 1 || group.GroupTypes[0] != CardTypeLand {
		t.Fatalf("group types = %v, want [CardTypeLand]", group.GroupTypes)
	}
	if !target.Exact {
		t.Fatal("target Exact = false, want true")
	}
}

// TestParseBareDestroyTargetHasNoSameNameGroup verifies that an ordinary destroy
// target without the clause does not spuriously gain a same-name group.
func TestParseBareDestroyTargetHasNoSameNameGroup(t *testing.T) {
	t.Parallel()
	targets := firstDestroyTargets(t, "Destroy target creature.")
	if len(targets) != 1 {
		t.Fatalf("targets = %d, want 1", len(targets))
	}
	if targets[0].Selection.SameNameGroup != nil {
		t.Fatal("bare target SameNameGroup != nil, want nil")
	}
}
