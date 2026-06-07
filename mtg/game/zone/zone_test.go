package zone

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/id"
)

func TestShuffleWithSameSeedProducesSameOrder(t *testing.T) {
	first := New(Library)
	second := New(Library)
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

func TestShufflePanicsOnNilRand(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("Shuffle(nil) did not panic")
		}
	}()

	cards := New(Library)
	cards.Shuffle(nil)
}

func TestRangeVisitsCardsInOrderAndCanStop(t *testing.T) {
	cards := New(Library)
	cards.AddToBottom(1)
	cards.AddToBottom(2)
	cards.AddToBottom(3)

	var visited []id.ID
	cards.Range(func(cardID id.ID) bool {
		visited = append(visited, cardID)
		return cardID != 2
	})

	if !slices.Equal(visited, []id.ID{1, 2}) {
		t.Fatalf("visited cards = %v, want [1 2]", visited)
	}
}
