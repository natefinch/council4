package action

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/id"
)

func TestPass(t *testing.T) {
	got := Pass()
	if got.Kind != ActionPass {
		t.Fatalf("Pass() kind = %v, want %v", got.Kind, ActionPass)
	}
}

func TestPlayLand(t *testing.T) {
	cardID := id.ID(42)

	got := PlayLand(cardID)
	if got.Kind != ActionPlayLand {
		t.Fatalf("PlayLand() kind = %v, want %v", got.Kind, ActionPlayLand)
	}
	if got.PlayLand.CardID != cardID {
		t.Fatalf("PlayLand() card ID = %v, want %v", got.PlayLand.CardID, cardID)
	}
}
