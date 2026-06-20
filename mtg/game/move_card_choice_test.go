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
			name: "bottom",
			move: MoveCard{
				Player:            ControllerReference(),
				Amount:            Fixed(2),
				FromZone:          zone.Hand,
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
