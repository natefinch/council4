package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerNarsetsReversalCopyThenReturn proves the copy-then-return stack-spell
// sequence lowers as an ordered pair sharing one stack-spell target: first a
// CopyStackObject that copies the targeted instant or sorcery spell (with the
// "you may choose new targets for the copy" rider), then a Bounce that returns
// the same targeted spell to its owner's hand. The two clauses reference the
// single "target instant or sorcery spell" slot through distinct object
// reference kinds — the copy addresses the stack object it duplicates, and the
// return addresses the target object it removes — so the remap walker must
// rebase both onto the shared target index 0.
func TestLowerNarsetsReversalCopyThenReturn(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Narset's Reversal",
		TypeLine: "Instant",
		OracleText: "Copy target instant or sorcery spell, then return it to its owner's hand. " +
			"You may choose new targets for the copy.",
	})

	mode := face.SpellAbility.Val.Modes[0]

	if len(mode.Targets) != 1 {
		t.Fatalf("target specs = %d, want 1", len(mode.Targets))
	}
	target := mode.Targets[0]
	if target.MinTargets != 1 || target.MaxTargets != 1 {
		t.Fatalf("target cardinality = %d..%d, want 1..1", target.MinTargets, target.MaxTargets)
	}
	if target.Allow != game.TargetAllowStackObject {
		t.Fatalf("target Allow = %v, want TargetAllowStackObject", target.Allow)
	}
	if kinds := target.Predicate.StackObjectKinds; len(kinds) != 1 || kinds[0] != game.StackSpell {
		t.Fatalf("StackObjectKinds = %v, want [StackSpell]", kinds)
	}
	wantTypes := []types.Card{types.Instant, types.Sorcery}
	if got := target.Predicate.SpellCardTypesAny; !equalCardTypes(got, wantTypes) {
		t.Fatalf("SpellCardTypesAny = %v, want %v", got, wantTypes)
	}

	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence length = %d, want 2 (copy then return)", len(mode.Sequence))
	}

	copyPrim, ok := mode.Sequence[0].Primitive.(game.CopyStackObject)
	if !ok {
		t.Fatalf("sequence[0] = %#v, want CopyStackObject", mode.Sequence[0].Primitive)
	}
	if !copyPrim.MayChooseNewTargets {
		t.Fatal("copy clause missing MayChooseNewTargets rider")
	}
	if copyPrim.Object.Kind() != game.ObjectReferenceTargetStackObject {
		t.Fatalf("copy object kind = %v, want TargetStackObject", copyPrim.Object.Kind())
	}
	if copyPrim.Object.TargetIndex() != 0 {
		t.Fatalf("copy object target index = %d, want 0", copyPrim.Object.TargetIndex())
	}

	bounce, ok := mode.Sequence[1].Primitive.(game.Bounce)
	if !ok {
		t.Fatalf("sequence[1] = %#v, want Bounce", mode.Sequence[1].Primitive)
	}
	if bounce.Object.Kind() != game.ObjectReferenceTargetObject {
		t.Fatalf("return object kind = %v, want TargetObject", bounce.Object.Kind())
	}
	if bounce.Object.TargetIndex() != 0 {
		t.Fatalf("return object target index = %d, want 0", bounce.Object.TargetIndex())
	}
}

// TestLowerReturnTargetInstantOrSorcerySpell proves the standalone spell-bounce
// form with a card-type filter ("Return target instant or sorcery spell to its
// owner's hand.") lowers to a single Bounce over a stack-spell target whose
// predicate restricts the legal targets to instants and sorceries. This guards
// the spellBounceTargetSpec change independently of the copy-then-return
// sequence so the type filter keeps working for any future card that only
// returns a typed spell.
func TestLowerReturnTargetInstantOrSorcerySpell(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Typed Spell Bounce",
		TypeLine:   "Instant",
		OracleText: "Return target instant or sorcery spell to its owner's hand.",
	})

	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("target specs = %d, want 1", len(mode.Targets))
	}
	target := mode.Targets[0]
	if target.Allow != game.TargetAllowStackObject {
		t.Fatalf("target Allow = %v, want TargetAllowStackObject", target.Allow)
	}
	if kinds := target.Predicate.StackObjectKinds; len(kinds) != 1 || kinds[0] != game.StackSpell {
		t.Fatalf("StackObjectKinds = %v, want [StackSpell]", kinds)
	}
	wantTypes := []types.Card{types.Instant, types.Sorcery}
	if got := target.Predicate.SpellCardTypesAny; !equalCardTypes(got, wantTypes) {
		t.Fatalf("SpellCardTypesAny = %v, want %v", got, wantTypes)
	}

	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1 (return only)", len(mode.Sequence))
	}
	bounce, ok := mode.Sequence[0].Primitive.(game.Bounce)
	if !ok {
		t.Fatalf("sequence[0] = %#v, want Bounce", mode.Sequence[0].Primitive)
	}
	if bounce.Object.TargetIndex() != 0 {
		t.Fatalf("return object target index = %d, want 0", bounce.Object.TargetIndex())
	}
}

func equalCardTypes(got, want []types.Card) bool {
	return slices.Equal(got, want)
}
