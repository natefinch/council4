package game

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCastForFreeXBoundValidation(t *testing.T) {
	primitive := CastForFree{
		Player:            ControllerReference(),
		Zone:              zone.Hand,
		MaxManaValueFromX: true,
	}
	if err := primitive.validatePrimitive(nil, false); err != nil {
		t.Fatalf("selection-driven X bound validation = %v", err)
	}
	primitive = CastForFree{
		Player:            ControllerReference(),
		Zone:              zone.Graveyard,
		Card:              CardReference{Kind: CardReferenceTarget},
		MaxManaValueFromX: true,
	}
	err := primitive.validatePrimitive(nil, false)
	if err == nil || !strings.Contains(err.Error(), "selection-driven") {
		t.Fatalf("referenced-card X bound error = %v", err)
	}
}
