package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerAvishkarRacewayMaxSpeed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Avishkar Raceway",
		Layout:   "normal",
		TypeLine: "Land",
		OracleText: "Start your engines! (If you have no speed, it starts at 1. It increases once on each of your turns when an opponent loses life. Max speed is 4.)\n" +
			"{T}: Add {C}.\n" +
			"Max speed — {3}, {T}, Discard a card: Draw a card.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want one Max speed ability", len(face.ActivatedAbilities))
	}
	ability := face.ActivatedAbilities[0]
	if ability.Timing != game.NoTimingRestriction {
		t.Fatalf("timing = %v, want NoTimingRestriction (Max speed imposes no timing)", ability.Timing)
	}
	if !ability.ActivationCondition.Exists || !ability.ActivationCondition.Val.ControllerHasMaxSpeed {
		t.Fatalf("activation condition = %#v, want ControllerHasMaxSpeed", ability.ActivationCondition)
	}
}

func TestLowerMaxSpeedRejectsExplicitTimingOrCondition(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Malformed Racer",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "Max speed — {3}, {T}: Draw a card. Activate only as a sorcery.",
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected a diagnostic for a Max speed ability with an explicit timing restriction")
	}
}
