package game

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/id"
)

func TestZoneShuffleWithSameSeedProducesSameOrder(t *testing.T) {
	first := NewZone(ZoneLibrary)
	second := NewZone(ZoneLibrary)
	for i := id.ID(1); i <= 10; i++ {
		first.AddToBottom(i)
		second.AddToBottom(i)
	}

	first.Shuffle(rand.New(rand.NewPCG(1, 2)))
	second.Shuffle(rand.New(rand.NewPCG(1, 2)))

	if !slices.Equal(first.All(), second.All()) {
		t.Fatalf("shuffle orders differ: %v != %v", first.All(), second.All())
	}
}

func TestZoneShufflePanicsOnNilRand(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Shuffle(nil) did not panic")
		}
	}()

	zone := NewZone(ZoneLibrary)
	zone.Shuffle(nil)
}
