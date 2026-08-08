package game

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

func TestValidateChosenHandToLibraryMove(t *testing.T) {
	t.Parallel()
	valid := MoveCard{
		Player:      ControllerReference(),
		Amount:      Fixed(2),
		FromZone:    zone.Hand,
		Destination: zone.Library,
	}
	if err := ValidateInstructionSequence([]Instruction{{Primitive: valid}}); err != nil {
		t.Fatalf("valid chosen-card move: %v", err)
	}
	// A chosen-card hand-to-library move may also request bottom placement
	// (handleMoveChosenHandCards reads DestinationBottom): Sawtooth Loon,
	// "draw two cards, then put two cards from your hand on the bottom of
	// your library."
	validBottom := MoveCard{
		Player:            ControllerReference(),
		Amount:            Fixed(2),
		FromZone:          zone.Hand,
		Destination:       zone.Library,
		DestinationBottom: true,
	}
	if err := ValidateInstructionSequence([]Instruction{{Primitive: validBottom}}); err != nil {
		t.Fatalf("valid chosen-card bottom move: %v", err)
	}

	tests := []struct {
		name string
		move MoveCard
		want string
	}{
		{
			name: "single card amount",
			move: MoveCard{
				Card:        CardReference{Kind: CardReferenceEvent},
				Amount:      Fixed(1),
				FromZone:    zone.Hand,
				Destination: zone.Library,
			},
			want: "single-card move must not set Amount",
		},
		{
			name: "wrong source",
			move: MoveCard{
				Player:      ControllerReference(),
				Amount:      Fixed(2),
				FromZone:    zone.Graveyard,
				Destination: zone.Library,
			},
			want: "chosen-card move requires hand to library",
		},
		{
			name: "wrong destination",
			move: MoveCard{
				Player:      ControllerReference(),
				Amount:      Fixed(2),
				FromZone:    zone.Hand,
				Destination: zone.Exile,
			},
			want: "chosen-card move requires hand to library",
		},
		{
			// The whole-zone move ("move every card from a player's zone",
			// Amount == 0) does not support bottom placement --
			// handleMoveCardZoneGroup's whole-zone branch always places on
			// top -- unlike the chosen-card move validated above.
			name: "whole-zone move bottom",
			move: MoveCard{
				Player:            ControllerReference(),
				FromZone:          zone.Graveyard,
				Destination:       zone.Library,
				DestinationBottom: true,
			},
			want: "must not request bottom placement",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateInstructionSequence([]Instruction{{Primitive: test.move}})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}
