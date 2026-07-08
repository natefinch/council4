package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerMonstrosityActivatedAbility proves the monstrosity keyword action,
// written out as an activated ability ("{cost}: Monstrosity N."), lowers to a
// game.Monstrosity primitive scoped to the source permanent and carrying the
// fixed counter count. The printed reminder text ("(If this creature isn't
// monstrous, put N +1/+1 counters on it and it becomes monstrous.)") is subsumed
// by the runtime guard and adds nothing to the lowered sequence.
func TestLowerMonstrosityActivatedAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Monstrous Creature",
		Layout:   "normal",
		TypeLine: "Creature — Test",
		OracleText: "{4}{G}{G}: Monstrosity 3. " +
			"(If this creature isn't monstrous, put three +1/+1 counters on it and it becomes monstrous.)",
		Power:     new("4"),
		Toughness: new("4"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("instruction count = %d, want 1", len(mode.Sequence))
	}
	monstrosity, ok := mode.Sequence[0].Primitive.(game.Monstrosity)
	if !ok {
		t.Fatalf("primitive = %#v, want game.Monstrosity", mode.Sequence[0].Primitive)
	}
	if monstrosity.Object != game.SourcePermanentReference() {
		t.Fatalf("monstrosity object = %#v, want source permanent", monstrosity.Object)
	}
	if monstrosity.Amount != game.Fixed(3) {
		t.Fatalf("monstrosity amount = %#v, want Fixed(3)", monstrosity.Amount)
	}
}
