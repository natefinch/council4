package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestSubtypeRiderKeywordGrant verifies that a pump-then-keyword sequence whose
// keyword grant is gated by a subtype rider ("if it's a <subtype>") lowers the
// gate as an ObjectMatches effect condition bound to the buffed target.
func TestSubtypeRiderKeywordGrant(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Subtype Rider",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gets +3/+2 until end of turn. If it's a Knight, it also gains first strike until end of turn.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d, want pump then gated keyword grant", len(mode.Sequence))
	}
	if _, ok := mode.Sequence[0].Primitive.(game.ModifyPT); !ok {
		t.Fatalf("sequence[0] = %T, want game.ModifyPT", mode.Sequence[0].Primitive)
	}
	grant, ok := mode.Sequence[1].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("sequence[1] = %T, want game.ApplyContinuous", mode.Sequence[1].Primitive)
	}
	if len(grant.ContinuousEffects) != 1 || len(grant.ContinuousEffects[0].AddKeywords) != 1 ||
		grant.ContinuousEffects[0].AddKeywords[0] != game.FirstStrike {
		t.Fatalf("grant effects = %+v, want first strike", grant.ContinuousEffects)
	}
	condition := requireSubtypeGate(t, mode.Sequence[1])
	if condition.Negate {
		t.Fatal("condition negated = true, want positive subtype gate")
	}
	if len(condition.ObjectMatches.Val.SubtypesAny) != 1 ||
		condition.ObjectMatches.Val.SubtypesAny[0] != types.Sub("Knight") {
		t.Fatalf("gate selection = %+v, want subtype Knight", condition.ObjectMatches.Val)
	}
}

// TestLegendaryRiderKeywordGrant verifies that a pump-then-keyword sequence gated
// by "if it's legendary" lowers the gate to a Legendary supertype match bound to
// the buffed target.
func TestLegendaryRiderKeywordGrant(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Legendary Rider",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gets +2/+2 until end of turn. If it's legendary, it also gains trample until end of turn.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d, want pump then gated keyword grant", len(mode.Sequence))
	}
	condition := requireSubtypeGate(t, mode.Sequence[1])
	if condition.Negate {
		t.Fatal("condition negated = true, want positive supertype gate")
	}
	if len(condition.ObjectMatches.Val.Supertypes) != 1 ||
		condition.ObjectMatches.Val.Supertypes[0] != types.Legendary {
		t.Fatalf("gate selection = %+v, want supertype Legendary", condition.ObjectMatches.Val)
	}
}

// TestSubtypeRiderInsteadPTKeywordGrant verifies that an "instead" conditional
// clause combining a PT change and keyword grant ("if it's a <subtype>, instead
// it gets +N/+N and gains <keyword>") gates the base buff on the negation and the
// replacement clauses on the positive subtype match, so exactly one branch runs.
func TestSubtypeRiderInsteadPTKeywordGrant(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Instead PT Keyword",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Target creature gets +2/+2 until end of turn. If it's a Human, instead it gets +3/+3 and gains indestructible until end of turn.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 3 {
		t.Fatalf("sequence = %d, want base buff plus replacement PT and keyword", len(mode.Sequence))
	}
	base := requireSubtypeGate(t, mode.Sequence[0])
	if !base.Negate {
		t.Fatalf("base buff gate = %+v, want negated Human match", base)
	}
	for i := 1; i < 3; i++ {
		gate := requireSubtypeGate(t, mode.Sequence[i])
		if gate.Negate {
			t.Fatalf("replacement[%d] gate negated = true, want positive Human match", i)
		}
		if len(gate.ObjectMatches.Val.SubtypesAny) != 1 ||
			gate.ObjectMatches.Val.SubtypesAny[0] != types.Sub("Human") {
			t.Fatalf("replacement[%d] selection = %+v, want subtype Human", i, gate.ObjectMatches.Val)
		}
	}
}

// requireSubtypeGate asserts the instruction carries an effect condition that
// matches its targeted permanent (the rider's "it") and returns the wrapped
// Condition for further assertions.
func requireSubtypeGate(t *testing.T, instruction game.Instruction) game.Condition {
	t.Helper()
	if !instruction.Condition.Exists {
		t.Fatalf("instruction %+v carries no effect condition gate", instruction.Primitive)
	}
	condition := instruction.Condition.Val.Condition
	if !condition.Exists {
		t.Fatal("effect condition carries no wrapped match condition")
	}
	if !condition.Val.Object.Exists ||
		condition.Val.Object.Val != game.TargetPermanentReference(0) {
		t.Fatalf("gate object = %+v, want target permanent 0", condition.Val.Object)
	}
	if !condition.Val.ObjectMatches.Exists {
		t.Fatalf("gate condition = %+v, want an ObjectMatches selection", condition.Val)
	}
	return condition.Val
}
