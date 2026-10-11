package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLibraryCardPlacementValidation(t *testing.T) {
	for _, test := range []struct {
		name string
		move MoveCard
		ok   bool
	}{
		{"top", MoveCard{Card: CardReference{Kind: CardReferenceLinked, LinkID: "observed"}, FromZone: zone.Library, Destination: zone.Library}, true},
		{"bottom", MoveCard{Card: CardReference{Kind: CardReferenceLinked, LinkID: "observed"}, FromZone: zone.Library, Destination: zone.Library, DestinationBottom: true}, true},
		{"missing identity", MoveCard{Card: CardReference{Kind: CardReferenceLinked}, FromZone: zone.Library, Destination: zone.Library}, false},
		{"hand same zone", MoveCard{Card: CardReference{Kind: CardReferenceLinked, LinkID: "observed"}, FromZone: zone.Hand, Destination: zone.Hand}, false},
		{"player same zone", MoveCard{Player: ControllerReference(), Amount: Fixed(1), FromZone: zone.Library, Destination: zone.Library}, false},
		{"bottom outside library", MoveCard{Card: CardReference{Kind: CardReferenceLinked, LinkID: "observed"}, FromZone: zone.Library, Destination: zone.Hand, DestinationBottom: true}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := test.move.validatePrimitive(nil, false); (err == nil) != test.ok {
				t.Fatalf("validation = %v, want accepted %v", err, test.ok)
			}
		})
	}
}
