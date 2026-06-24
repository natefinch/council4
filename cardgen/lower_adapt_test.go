package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerAdaptActivatedAbility proves the Adapt keyword action, written out as
// an activated ability ("{cost}: Adapt N."), lowers to a game.Adapt primitive
// scoped to the source permanent and carrying the fixed counter count. The
// printed reminder text ("(If this creature has no +1/+1 counters on it, put N
// +1/+1 counters on it.)") is subsumed by the runtime guard and adds nothing to
// the lowered sequence.
func TestLowerAdaptActivatedAbility(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Adapt Creature",
		Layout:   "normal",
		TypeLine: "Creature — Test",
		OracleText: "{4}{G}{G}: Adapt 4. " +
			"(If this creature has no +1/+1 counters on it, put four +1/+1 counters on it.)",
		Power:     new("3"),
		Toughness: new("3"),
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("instruction count = %d, want 1", len(mode.Sequence))
	}
	adapt, ok := mode.Sequence[0].Primitive.(game.Adapt)
	if !ok {
		t.Fatalf("primitive = %#v, want game.Adapt", mode.Sequence[0].Primitive)
	}
	if adapt.Object != game.SourcePermanentReference() {
		t.Fatalf("adapt object = %#v, want source permanent", adapt.Object)
	}
	if adapt.Amount != game.Fixed(4) {
		t.Fatalf("adapt amount = %#v, want Fixed(4)", adapt.Amount)
	}
}
