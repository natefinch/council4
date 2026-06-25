package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func eachPlayerBasicLandSearchSpec() SearchSpec {
	return SearchSpec{
		SourceZone:  zone.Library,
		Destination: zone.Battlefield,
		Filter: Selection{
			RequiredTypes: []types.Card{types.Land},
			Supertypes:    []types.Super{types.Basic},
		},
	}
}

func TestSearchPlayerGroupValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		primitive Search
		wantError bool
	}{
		{
			name: "each player group",
			primitive: Search{
				PlayerGroup: AllPlayersReference(),
				Amount:      Fixed(1),
				Spec:        eachPlayerBasicLandSearchSpec(),
			},
		},
		{
			name: "both player and group set",
			primitive: Search{
				Player:      ControllerReference(),
				PlayerGroup: AllPlayersReference(),
				Amount:      Fixed(1),
				Spec:        eachPlayerBasicLandSearchSpec(),
			},
			wantError: true,
		},
		{
			name: "group with controller",
			primitive: Search{
				PlayerGroup: AllPlayersReference(),
				Controller:  opt.Val(ControllerReference()),
				Amount:      Fixed(1),
				Spec:        eachPlayerBasicLandSearchSpec(),
			},
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateInstructionSequence([]Instruction{{Primitive: tt.primitive}}, nil)
			if (err != nil) != tt.wantError {
				t.Fatalf("ValidateInstructionSequence() error = %v, wantError = %v", err, tt.wantError)
			}
		})
	}
}
