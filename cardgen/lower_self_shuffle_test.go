package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// selfShuffleInstruction lowers a single dies / put-into-graveyard triggered
// ability and returns its first instruction.
func selfShuffleInstruction(t *testing.T, oracleText string) (game.Instruction, bool) {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Self Shuffle",
		Layout:     "normal",
		TypeLine:   "Creature — Dragon",
		OracleText: oracleText,
	})
	if len(face.TriggeredAbilities) == 0 {
		t.Fatal("no triggered abilities")
	}
	ab := face.TriggeredAbilities[0]
	return ab.Content.Modes[0].Sequence[0], ab.Optional
}

// TestLowerSelfShuffleIntoOwnerLibrary lowers "When this creature dies, shuffle
// it into its owner's library." to a ShufflePermanentIntoLibrary naming the
// triggering permanent, mandatory at the instruction level.
func TestLowerSelfShuffleIntoOwnerLibrary(t *testing.T) {
	t.Parallel()
	instr, abilityOptional := selfShuffleInstruction(t,
		"When this creature dies, shuffle it into its owner's library.")
	shuffle, ok := instr.Primitive.(game.ShufflePermanentIntoLibrary)
	if !ok {
		t.Fatalf("primitive = %T, want game.ShufflePermanentIntoLibrary", instr.Primitive)
	}
	if shuffle.Object != game.EventPermanentReference() {
		t.Fatalf("object = %#v, want event permanent reference", shuffle.Object)
	}
	if instr.Optional || abilityOptional {
		t.Fatalf("mandatory shuffle lowered optional (instr=%v ability=%v)", instr.Optional, abilityOptional)
	}
}

// TestLowerSelfShuffleOptionalKeepsAbilityOptional lowers the "you may" form and
// confirms the optionality is preserved at the ability level (a mandatory
// instruction under a "you may" ability), not dropped.
func TestLowerSelfShuffleOptionalKeepsAbilityOptional(t *testing.T) {
	t.Parallel()
	instr, abilityOptional := selfShuffleInstruction(t,
		"When this creature dies, you may shuffle it into its owner's library.")
	if _, ok := instr.Primitive.(game.ShufflePermanentIntoLibrary); !ok {
		t.Fatalf("primitive = %T, want game.ShufflePermanentIntoLibrary", instr.Primitive)
	}
	if !abilityOptional {
		t.Fatal("optional \"you may\" shuffle lost its ability-level optionality")
	}
}
