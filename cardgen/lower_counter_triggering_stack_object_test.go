package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerCounterTriggeringSpellOrAbility verifies a became-target trigger's
// "counter that spell or ability" body counters the triggering stack object
// without announcing a new target.
func TestLowerCounterTriggeringSpellOrAbility(t *testing.T) {
	t.Parallel()
	const oracleText = "Whenever this creature becomes the target of a spell or ability for the first time each turn, counter that spell or ability."
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Shimmering Glasskite",
		Layout:     "normal",
		TypeLine:   "Creature — Spirit",
		OracleText: "Flying\n" + oracleText,
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %#v, want none", mode.Targets)
	}
	counter, ok := mode.Sequence[0].Primitive.(game.CounterObject)
	if !ok || counter.Object != game.EventStackObjectReference() {
		t.Fatalf("primitive = %#v, want CounterObject of event stack object", mode.Sequence[0].Primitive)
	}
}
