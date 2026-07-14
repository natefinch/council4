package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestDigLibraryTopDestinationValidates proves a Dig that keeps up to one chosen
// card on top of the library (Thassa's Oracle) is a valid primitive: a dynamic
// look count, a fixed take of one, the library-bottom remainder, and a
// library-top destination.
func TestDigLibraryTopDestinationValidates(t *testing.T) {
	t.Parallel()
	dig := Dig{
		Player: ControllerReference(),
		Look: Dynamic(DynamicAmount{
			Kind:   DynamicAmountDevotion,
			Colors: []color.Color{color.Blue},
		}),
		Take:        Fixed(1),
		TakeUpTo:    true,
		Destination: zone.Library,
		Remainder:   DigRemainderLibraryBottom,
	}
	if err := dig.validatePrimitive(nil, true); err != nil {
		t.Fatalf("library-top dig validation failed: %v", err)
	}
}

// TestDigLibraryTopDestinationRejectsReveal proves a library-top dig cannot also
// reveal the chosen cards: keeping a card on top of the library is a hidden
// action, so combining zone.Library with Reveal is rejected.
func TestDigLibraryTopDestinationRejectsReveal(t *testing.T) {
	t.Parallel()
	dig := Dig{
		Player:      ControllerReference(),
		Look:        Fixed(3),
		Take:        Fixed(1),
		Destination: zone.Library,
		Remainder:   DigRemainderLibraryBottom,
		Reveal:      true,
	}
	if err := dig.validatePrimitive(nil, true); err == nil {
		t.Fatal("library-top dig with Reveal validated, want rejection")
	}
}
