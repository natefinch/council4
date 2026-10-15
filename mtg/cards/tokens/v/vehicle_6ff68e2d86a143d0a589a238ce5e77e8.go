package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Vehicle is the card definition for Vehicle.
//
// Type: Token Artifact — Vehicle
//
// Oracle text:
//   Crew 1 (Tap any number of creatures you control with total power 1 or more: This token becomes an artifact creature until end of turn.)

// VehicleToken6ff68e2d86a143d0a589a238ce5e77e8 is the card definition for Vehicle.
var VehicleToken6ff68e2d86a143d0a589a238ce5e77e8 = newVehicleToken6ff68e2d86a143d0a589a238ce5e77e8()

func newVehicleToken6ff68e2d86a143d0a589a238ce5e77e8() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Vehicle",
			Types:     []types.Card{types.Artifact},
			Subtypes:  []types.Sub{types.Vehicle},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.CrewActivatedAbility(1),
			},
			OracleText: `
			Crew 1 (Tap any number of creatures you control with total power 1 or more: This token becomes an artifact creature until end of turn.)
		`,
		},
	}
}
