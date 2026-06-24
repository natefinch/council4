package game

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TestChooseFromZoneFamilyValidation pins the accept/reject set of every
// choose-from-zone-family primitive after they were consolidated onto the single
// validateZoneChoice validator. Each case proves that the canonical validator
// reproduces the historical per-family restriction exactly, including the
// divergences the families must keep (only the object-scoped imprint exile
// requires a single card; only the graveyard-return family validates a
// destination).
func TestChooseFromZoneFamilyValidation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		prim    Primitive
		wantErr string // empty means the primitive must validate
	}{
		{
			name: "exile from hand accepts a fixed positive amount",
			prim: ExileFromHand{Player: ControllerReference(), Amount: Fixed(1)},
		},
		{
			name:    "exile from hand rejects a non-positive amount",
			prim:    ExileFromHand{Player: ControllerReference(), Amount: Fixed(0)},
			wantErr: "exile from hand requires a fixed positive amount",
		},
		{
			name: "exile from hand accepts a linked single-card exile",
			prim: ExileFromHand{Player: ControllerReference(), Amount: Fixed(1), PublishLinked: "imprint"},
		},
		{
			name:    "exile from hand rejects a linked multi-card exile",
			prim:    ExileFromHand{Player: ControllerReference(), Amount: Fixed(2), PublishLinked: "imprint"},
			wantErr: "linked exile from hand must exile exactly one card",
		},
		{
			name: "exile from graveyard accepts a fixed positive amount",
			prim: ExileFromGraveyard{Player: ControllerReference(), Amount: Fixed(1)},
		},
		{
			// The card-scoped graveyard publish is not single-card gated, unlike
			// the object-scoped hand imprint above.
			name: "exile from graveyard accepts a linked multi-card exile",
			prim: ExileFromGraveyard{Player: ControllerReference(), Amount: Fixed(2), PublishLinked: "set"},
		},
		{
			name:    "exile from graveyard rejects a non-positive amount",
			prim:    ExileFromGraveyard{Player: ControllerReference(), Amount: Fixed(0)},
			wantErr: "exile from graveyard requires a fixed positive amount",
		},
		{
			name: "put from hand accepts a fixed positive amount",
			prim: PutFromHand{Player: ControllerReference(), Amount: Fixed(1)},
		},
		{
			name:    "put from hand rejects a non-positive amount",
			prim:    PutFromHand{Player: ControllerReference(), Amount: Fixed(0)},
			wantErr: "put from hand requires a fixed positive amount",
		},
		{
			name: "return from graveyard accepts a hand return",
			prim: ReturnFromGraveyard{Player: ControllerReference(), Amount: Fixed(1), Destination: zone.Hand},
		},
		{
			name: "return from graveyard accepts a default destination",
			prim: ReturnFromGraveyard{Player: ControllerReference(), Amount: Fixed(1)},
		},
		{
			name: "return from graveyard accepts a tapped battlefield return",
			prim: ReturnFromGraveyard{Player: ControllerReference(), Amount: Fixed(1), Destination: zone.Battlefield, EntryTapped: true},
		},
		{
			name: "return from graveyard accepts an any-number form",
			prim: ReturnFromGraveyard{Player: ControllerReference(), AnyNumber: true},
		},
		{
			name: "return from graveyard accepts a capped battlefield return",
			prim: ReturnFromGraveyard{Player: ControllerReference(), Amount: Fixed(2), Destination: zone.Battlefield, MaxTotalManaValue: opt.Val(4)},
		},
		{
			name:    "return from graveyard rejects a non-positive amount",
			prim:    ReturnFromGraveyard{Player: ControllerReference(), Amount: Fixed(0)},
			wantErr: "return from graveyard requires a fixed positive amount",
		},
		{
			name:    "return from graveyard rejects an any-number form with a fixed amount",
			prim:    ReturnFromGraveyard{Player: ControllerReference(), AnyNumber: true, Amount: Fixed(2)},
			wantErr: "return from graveyard any-number form takes no fixed amount",
		},
		{
			name:    "return from graveyard rejects an any-number form with a cap",
			prim:    ReturnFromGraveyard{Player: ControllerReference(), AnyNumber: true, MaxTotalManaValue: opt.Val(4)},
			wantErr: "return from graveyard any-number form takes no total mana value cap",
		},
		{
			name:    "return from graveyard rejects an unsupported destination",
			prim:    ReturnFromGraveyard{Player: ControllerReference(), Amount: Fixed(1), Destination: zone.Library},
			wantErr: "return from graveyard requires a hand or battlefield destination",
		},
		{
			name:    "return from graveyard rejects a tapped non-battlefield return",
			prim:    ReturnFromGraveyard{Player: ControllerReference(), Amount: Fixed(1), Destination: zone.Hand, EntryTapped: true},
			wantErr: "return from graveyard tapped entry requires a battlefield destination",
		},
		{
			name:    "return from graveyard rejects a capped non-battlefield return",
			prim:    ReturnFromGraveyard{Player: ControllerReference(), Amount: Fixed(1), Destination: zone.Hand, MaxTotalManaValue: opt.Val(4)},
			wantErr: "return from graveyard total mana value cap requires a battlefield destination",
		},
		{
			name:    "return from graveyard rejects a negative cap",
			prim:    ReturnFromGraveyard{Player: ControllerReference(), Amount: Fixed(1), Destination: zone.Battlefield, MaxTotalManaValue: opt.Val(-1)},
			wantErr: "return from graveyard total mana value cap must be non-negative",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateInstructionSequence([]Instruction{{Primitive: tc.prim}}, nil)
			switch {
			case tc.wantErr == "" && err != nil:
				t.Fatalf("ValidateInstructionSequence() = %v, want nil", err)
			case tc.wantErr != "" && err == nil:
				t.Fatalf("ValidateInstructionSequence() = nil, want error %q", tc.wantErr)
			case tc.wantErr != "" && !strings.Contains(err.Error(), tc.wantErr):
				t.Fatalf("ValidateInstructionSequence() = %v, want error containing %q", err, tc.wantErr)
			default:
			}
		})
	}
}
