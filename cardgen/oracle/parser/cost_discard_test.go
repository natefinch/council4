package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

func TestParseDiscardWholeHandCost(t *testing.T) {
	t.Parallel()
	component := soleCostComponent(t, "Discard your hand: Add {C}.")
	if component.Kind != CostComponentDiscard {
		t.Fatalf("kind = %v, want discard", component.Kind)
	}
	if !component.DiscardWholeHand {
		t.Fatal("DiscardWholeHand = false, want true")
	}
	if component.SourceZone != zone.Hand {
		t.Fatalf("source zone = %v, want hand", component.SourceZone)
	}
	if component.AmountKnown {
		t.Fatal("AmountKnown = true, want false for a whole-hand discard")
	}
}

func TestParseDiscardFixedCardCostIsNotWholeHand(t *testing.T) {
	t.Parallel()
	component := soleCostComponent(t, "Discard a card: Draw a card.")
	if component.Kind != CostComponentDiscard {
		t.Fatalf("kind = %v, want discard", component.Kind)
	}
	if component.DiscardWholeHand {
		t.Fatal("DiscardWholeHand = true, want false for a fixed-count discard")
	}
	if !component.AmountKnown || component.AmountValue != 1 {
		t.Fatalf("amount = (%d, %v), want 1", component.AmountValue, component.AmountKnown)
	}
}
