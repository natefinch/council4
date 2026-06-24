package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerConclaveMentorDiesGainLifeEqualToPower proves Conclave Mentor's "When
// this creature dies, you gain life equal to its power." lowers to a dies-
// triggered GainLife whose amount reads the dying permanent's last-known power
// through an event-permanent reference, the source-power amount form a leaves-
// the-battlefield trigger needs.
func TestLowerConclaveMentorDiesGainLifeEqualToPower(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:      "Conclave Mentor",
		Layout:    "normal",
		TypeLine:  "Creature — Centaur Cleric",
		ManaCost:  "{G}{W}",
		Power:     new("2"),
		Toughness: new("2"),
		OracleText: "If one or more +1/+1 counters would be put on a creature you control, that many plus one +1/+1 counters are put on that creature instead.\n" +
			"When this creature dies, you gain life equal to its power.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"game.CounterPlacementReplacement(",
		"counter.PlusOnePlusOne",
		"game.TriggerControllerYou",
		"game.EventPermanentDied",
		"Primitive: game.GainLife{",
		"game.DynamicAmountObjectPower",
		"game.EventPermanentReference()",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestLowerConclaveMentorGroupCounterReplacementAmount proves Conclave Mentor's
// group counter-amount replacement lowers to a "that many plus one" amount
// (multiplier 0, addend 1) scoped to creatures the controller controls.
func TestLowerConclaveMentorGroupCounterReplacementAmount(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:      "Conclave Mentor",
		Layout:    "normal",
		TypeLine:  "Creature — Centaur Cleric",
		ManaCost:  "{G}{W}",
		Power:     new("2"),
		Toughness: new("2"),
		OracleText: "If one or more +1/+1 counters would be put on a creature you control, that many plus one +1/+1 counters are put on that creature instead.\n" +
			"When this creature dies, you gain life equal to its power.",
	})
	var counterReplacement *game.ReplacementEffect
	for i := range face.ReplacementAbilities {
		if face.ReplacementAbilities[i].Replacement.CounterMultiplier != 0 ||
			face.ReplacementAbilities[i].Replacement.CounterAddend != 0 {
			counterReplacement = &face.ReplacementAbilities[i].Replacement
		}
	}
	if counterReplacement == nil {
		t.Fatalf("no counter-amount replacement lowered: %#v", face.ReplacementAbilities)
	}
	if counterReplacement.CounterMultiplier != 0 || counterReplacement.CounterAddend != 1 {
		t.Fatalf("counter amount = (mul %d, add %d), want (0, 1)", counterReplacement.CounterMultiplier, counterReplacement.CounterAddend)
	}
	if counterReplacement.CounterRecipientSelf {
		t.Fatalf("group replacement must not be self-scoped: %#v", counterReplacement)
	}
}
