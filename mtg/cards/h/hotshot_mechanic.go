package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HotshotMechanic is the card definition for Hotshot Mechanic.
//
// Type: Artifact Creature — Fox Pilot
// Cost: {W}
//
// Oracle text:
//
//	This creature crews Vehicles as though its power were 2 greater.
var HotshotMechanic = newHotshotMechanic

func newHotshotMechanic() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Hotshot Mechanic",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Fox, types.Pilot},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text:           "This creature crews Vehicles as though its power were 2 greater.",
					CrewPowerBonus: 2,
				},
			},
			OracleText: `
			This creature crews Vehicles as though its power were 2 greater.
		`,
		},
	}
}
