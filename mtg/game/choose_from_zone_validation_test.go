package game

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TestChooseFromZoneFamilyValidation pins the accept/reject set of the single
// canonical ChooseFromZone validator. Each case proves that the unified
// validator reproduces the historical per-family restriction exactly, including
// the divergences the families must keep: only the object-scoped imprint exile
// requires a single card, and only a battlefield destination admits a tapped
// entry or a total-mana-value cap. The envelopes mirror what the family
// constructors (ExileFromHandChoice, ReturnFromGraveyardChoice, ...) build.
func TestChooseFromZoneFamilyValidation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		prim    Primitive
		wantErr string // empty means the primitive must validate
	}{
		{
			name: "exile from hand accepts a fixed positive amount",
			prim: ExileFromHandChoice(ControllerReference(), Selection{}, Fixed(1), ""),
		},
		{
			name:    "exile from hand rejects a non-positive amount",
			prim:    ExileFromHandChoice(ControllerReference(), Selection{}, Fixed(0), ""),
			wantErr: "choose from zone requires a fixed positive amount",
		},
		{
			name: "exile from hand accepts a linked single-card exile",
			prim: ExileFromHandChoice(ControllerReference(), Selection{}, Fixed(1), "imprint"),
		},
		{
			name:    "exile from hand rejects a linked multi-card exile",
			prim:    ExileFromHandChoice(ControllerReference(), Selection{}, Fixed(2), "imprint"),
			wantErr: "linked choose from zone must move exactly one card",
		},
		{
			name: "exile from graveyard accepts a fixed positive amount",
			prim: ExileFromGraveyardChoice(ControllerReference(), Selection{}, Fixed(1), false, ""),
		},
		{
			// The card-scoped graveyard publish is not single-card gated, unlike
			// the object-scoped hand imprint above.
			name: "exile from graveyard accepts a linked multi-card exile",
			prim: ExileFromGraveyardChoice(ControllerReference(), Selection{}, Fixed(2), false, "set"),
		},
		{
			name:    "exile from graveyard rejects a non-positive amount",
			prim:    ExileFromGraveyardChoice(ControllerReference(), Selection{}, Fixed(0), false, ""),
			wantErr: "choose from zone requires a fixed positive amount",
		},
		{
			name: "put from hand accepts a fixed positive amount",
			prim: PutFromHandChoice(ControllerReference(), Selection{}, Fixed(1), false, false),
		},
		{
			name:    "put from hand rejects a non-positive amount",
			prim:    PutFromHandChoice(ControllerReference(), Selection{}, Fixed(0), false, false),
			wantErr: "choose from zone requires a fixed positive amount",
		},
		{
			name: "return from graveyard accepts a hand return",
			prim: ReturnFromGraveyardChoice(ControllerReference(), Selection{}, Fixed(1), zone.Hand, false, opt.V[int]{}, false, ""),
		},
		{
			name: "return from graveyard accepts a default destination",
			prim: ReturnFromGraveyardChoice(ControllerReference(), Selection{}, Fixed(1), zone.None, false, opt.V[int]{}, false, ""),
		},
		{
			name: "return from graveyard accepts a tapped battlefield return",
			prim: ReturnFromGraveyardChoice(ControllerReference(), Selection{}, Fixed(1), zone.Battlefield, true, opt.V[int]{}, false, ""),
		},
		{
			name: "return from graveyard accepts an any-number form",
			prim: ReturnFromGraveyardChoice(ControllerReference(), Selection{}, Quantity{}, zone.Battlefield, false, opt.V[int]{}, true, ""),
		},
		{
			name: "return from graveyard accepts a capped battlefield return",
			prim: ReturnFromGraveyardChoice(ControllerReference(), Selection{}, Fixed(2), zone.Battlefield, false, opt.Val(4), false, ""),
		},
		{
			name:    "return from graveyard rejects a non-positive amount",
			prim:    ReturnFromGraveyardChoice(ControllerReference(), Selection{}, Fixed(0), zone.Hand, false, opt.V[int]{}, false, ""),
			wantErr: "choose from zone requires a fixed positive amount",
		},
		{
			name:    "any-number form rejects a fixed amount",
			prim:    ChooseFromZone{Player: ControllerReference(), SourceZone: zone.Graveyard, Count: ChooseAnyNumber, Quantity: Fixed(2), Destination: ChooseDestination{Zone: zone.Battlefield}},
			wantErr: "choose from zone any-number form takes no fixed amount",
		},
		{
			name:    "any-number form rejects a total mana value cap",
			prim:    ChooseFromZone{Player: ControllerReference(), SourceZone: zone.Graveyard, Count: ChooseAnyNumber, Destination: ChooseDestination{Zone: zone.Battlefield}, Riders: ChooseRiders{MaxTotalManaValue: opt.Val(4)}},
			wantErr: "choose from zone any-number form takes no total mana value cap",
		},
		{
			name:    "choose from zone rejects a missing source zone",
			prim:    ChooseFromZone{Player: ControllerReference(), Quantity: Fixed(1), Destination: ChooseDestination{Zone: zone.Hand}},
			wantErr: "choose from zone requires a source zone",
		},
		{
			name:    "choose from zone rejects an unsupported destination",
			prim:    ChooseFromZone{Player: ControllerReference(), SourceZone: zone.Graveyard, Quantity: Fixed(1), Destination: ChooseDestination{Zone: zone.Library}},
			wantErr: "choose from zone requires an exile, hand, or battlefield destination",
		},
		{
			name:    "choose from zone rejects a tapped non-battlefield return",
			prim:    ChooseFromZone{Player: ControllerReference(), SourceZone: zone.Graveyard, Quantity: Fixed(1), Destination: ChooseDestination{Zone: zone.Hand}, Riders: ChooseRiders{EntersTapped: true}},
			wantErr: "choose from zone tapped entry requires a battlefield destination",
		},
		{
			name:    "choose from zone rejects a capped non-battlefield return",
			prim:    ChooseFromZone{Player: ControllerReference(), SourceZone: zone.Graveyard, Quantity: Fixed(1), Destination: ChooseDestination{Zone: zone.Hand}, Riders: ChooseRiders{MaxTotalManaValue: opt.Val(4)}},
			wantErr: "choose from zone total mana value cap requires a battlefield destination",
		},
		{
			name:    "choose from zone rejects a negative cap",
			prim:    ChooseFromZone{Player: ControllerReference(), SourceZone: zone.Graveyard, Quantity: Fixed(1), Destination: ChooseDestination{Zone: zone.Battlefield}, Riders: ChooseRiders{MaxTotalManaValue: opt.Val(-1)}},
			wantErr: "choose from zone total mana value cap must be non-negative",
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
