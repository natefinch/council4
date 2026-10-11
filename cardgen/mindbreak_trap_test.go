package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
)

const mindbreakTrapText = "If an opponent cast three or more spells this turn, you may pay {0} rather than pay this spell's mana cost.\n" +
	"Exile any number of target spells."

// TestLowerMindbreakTrap proves the full parser→compiler→lowering pipeline turns
// Mindbreak Trap's real Oracle text into both typed shapes issue #1779 adds: the
// per-opponent spells-cast {0} alternative cost and the variable-count
// exile-target-spells effect over an any-number stack-spell target group.
func TestLowerMindbreakTrap(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Mindbreak Trap",
		Layout:     "normal",
		TypeLine:   "Instant — Trap",
		ManaCost:   "{2}{U}{U}",
		OracleText: mindbreakTrapText,
	})

	if len(face.AlternativeCosts) != 1 {
		t.Fatalf("alternative costs = %#v, want 1", face.AlternativeCosts)
	}
	alt := face.AlternativeCosts[0]
	if alt.Condition != cost.AlternativeConditionOpponentCastSpellsThisTurn {
		t.Fatalf("alternative condition = %v, want OpponentCastSpellsThisTurn", alt.Condition)
	}
	if alt.ConditionCount != 3 {
		t.Fatalf("condition count = %d, want 3", alt.ConditionCount)
	}
	if alt.ConditionExactly {
		t.Fatal("condition is exact, want a 'three or more' minimum")
	}
	if !alt.ManaCost.Exists || alt.ManaCost.Val.String() != "{0}" {
		t.Fatalf("alternative mana cost = %#v, want {0}", alt.ManaCost)
	}

	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v, want one any-number spell target group", mode.Targets)
	}
	spec := mode.Targets[0]
	if spec.MinTargets != 0 || spec.MaxTargets != 99 {
		t.Fatalf("target cardinality = [%d,%d], want [0,99]", spec.MinTargets, spec.MaxTargets)
	}
	if spec.Allow&game.TargetAllowStackObject == 0 {
		t.Fatalf("target Allow = %v, want stack objects", spec.Allow)
	}
	if !slices.Equal(spec.Predicate.StackObjectKinds, []game.StackObjectKind{game.StackSpell}) {
		t.Fatalf("stack object kinds = %v, want [StackSpell]", spec.Predicate.StackObjectKinds)
	}

	exile, ok := mode.Sequence[0].Primitive.(game.ExileTargetSpells)
	if !ok {
		t.Fatalf("primitive = %#v, want game.ExileTargetSpells", mode.Sequence[0].Primitive)
	}
	if exile.Object != game.AllTargetStackObjectsReference(0) {
		t.Fatalf("exile object = %#v, want AllTargetStackObjectsReference(0)", exile.Object)
	}
}

func TestLowerAnyNumberPermanentExileUsesPermanentPrimitive(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Not Mindbreak",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{2}{U}",
		OracleText: "Exile any number of target creatures.",
	}
	assertCardPaths(t, card,
		"Modes[0].Sequence[0].Primitive.(game.MovePermanent).Object.kind = game.ObjectReferenceAllTargetPermanents",
		"Modes[0].Sequence[0].Primitive.(game.MovePermanent).Destination = zone.Exile",
	)
	assertCardPathsAbsent(t, card, "Primitive.(game.ExileTargetSpells)")
}

// TestLowerBoundedTargetSpellsExileFailsClosed proves the any-number exile path
// is limited to the unbounded "any number of" cardinality it was written for: a
// bounded "up to three target spells" is not silently mis-lowered as an
// unbounded group exile but fails closed, leaving room for a future bounded
// primitive.
func TestLowerBoundedTargetSpellsExileFailsClosed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Bounded Mindbreak",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{2}{U}",
		OracleText: "Exile up to three target spells.",
	})
	if face.SpellAbility.Exists {
		for _, mode := range face.SpellAbility.Val.Modes {
			for _, instr := range mode.Sequence {
				if _, ok := instr.Primitive.(game.ExileTargetSpells); ok {
					t.Fatal("bounded spell exile wrongly lowered to the unbounded ExileTargetSpells")
				}
			}
		}
	}
}
