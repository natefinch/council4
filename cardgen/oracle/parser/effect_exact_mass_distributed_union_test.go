package parser

import (
	"slices"
	"testing"
)

// TestExactDestroyMassDistributedUnionAccepts covers the distributed type-union
// wording of a mass group destroy, where the trailing controller clause repeats
// on every conjoined permanent noun ("Destroy all creatures you don't control and
// all planeswalkers you don't control.", In Garruk's Wake). It is semantically
// identical to the canonical shared-suffix union "creatures and planeswalkers you
// don't control" and must round-trip to the same exact group destroy.
func TestExactDestroyMassDistributedUnionAccepts(t *testing.T) {
	t.Parallel()
	accepted := []string{
		// In Garruk's Wake: two types under a repeated "you don't control".
		"Destroy all creatures you don't control and all planeswalkers you don't control.",
		// The same distributed shape under the other controller relations.
		"Destroy all creatures you control and all planeswalkers you control.",
		"Destroy all creatures and all planeswalkers.",
		// A noncreature union whose canonical form is also a recognized base noun.
		"Destroy all artifacts you don't control and all enchantments you don't control.",
	}
	for _, source := range accepted {
		if !destroyEffectExact(t, source) {
			t.Errorf("destroyEffectExact(%q) = false, want true", source)
		}
	}
}

// TestExactDestroyMassDistributedUnionFailsClosed proves the distributed-union
// recognition never fails open. The parser merges a mismatched-per-half controller
// into a single lossy relation; because the collapse requires an identical
// controller clause on every member, such a source is never reconstructed and must
// fail closed rather than destroy a wider set than the text names. A distributed
// form is also accepted only when its canonical equivalent would be, so a union
// noun with no canonical base ("creatures and artifacts") fails closed too.
func TestExactDestroyMassDistributedUnionFailsClosed(t *testing.T) {
	t.Parallel()
	rejected := []string{
		// Mismatched controllers on each half: the parser keeps only the first
		// half's relation, so accepting this would destroy the wrong permanents.
		"Destroy all creatures you control and all planeswalkers you don't control.",
		// The second half carries no controller clause at all.
		"Destroy all creatures you don't control and all planeswalkers.",
		// "creatures and artifacts" is not a canonical union base noun (only
		// "artifacts and creatures" is), so the distributed form fails closed in
		// lockstep with its canonical equivalent.
		"Destroy all creatures your opponents control and all artifacts your opponents control.",
	}
	for _, source := range rejected {
		if destroyEffectExact(t, source) {
			t.Errorf("destroyEffectExact(%q) = true, want false (fail closed)", source)
		}
	}
}

// TestExactDestroyMassDistributedUnionSelection checks that the In Garruk's Wake
// wording lowers to the correct typed selection: a Creature/Planeswalker type
// union scoped to permanents the caster does not control, selecting every match.
func TestExactDestroyMassDistributedUnionSelection(t *testing.T) {
	t.Parallel()
	const source = "Destroy all creatures you don't control and all planeswalkers you don't control."
	document, diagnostics := Parse(source, Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectDestroy {
		t.Fatalf("Parse(%q) effects = %#v", source, effects)
	}
	if !effects[0].Exact {
		t.Fatalf("destroyEffectExact(%q) = false, want true", source)
	}
	selection := effects[0].Selection
	if !selection.All {
		t.Error("selection.All = false, want true")
	}
	if selection.Controller != SelectionControllerNotYou {
		t.Errorf("selection.Controller = %v, want %v", selection.Controller, SelectionControllerNotYou)
	}
	wantTypes := []CardType{CardTypeCreature, CardTypePlaneswalker}
	if !slices.Equal(selection.RequiredTypesAny, wantTypes) {
		t.Errorf("selection.RequiredTypesAny = %#v, want %#v", selection.RequiredTypesAny, wantTypes)
	}
}
