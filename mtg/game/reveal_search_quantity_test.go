package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestRequiredRevealSearchValidatesExactUnrestrictedQuantity(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		count  int
		filter Selection
		names  bool
		valid  bool
	}{
		{name: "three unrestricted cards", count: 3, valid: true},
		{name: "one unrestricted card", count: 1, valid: true},
		{name: "no cards", count: 0},
		{name: "qualified cards", count: 3, filter: Selection{RequiredTypes: []types.Card{types.Land}}},
		{name: "different names", count: 3, names: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			operation := Search{
				Player: ControllerReference(), Amount: Fixed(tc.count),
				Spec: SearchSpec{
					SourceZone: zone.Library, Reveal: true, RevealOnly: true,
					Filter: tc.filter, DifferentNames: tc.names, FailToFindPolicy: SearchMustFindIfAvailable,
				},
			}
			if err := ValidatePrimitive(operation, nil); (err == nil) != tc.valid {
				t.Fatalf("validation = %v, valid = %v", err, tc.valid)
			}
		})
	}
}
