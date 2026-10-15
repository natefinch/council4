package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FleetfootDancer is the card definition for Fleetfoot Dancer.
//
// Type: Creature — Elf Druid
// Cost: {1}{R}{G}{W}
//
// Oracle text:
//
//	Trample, lifelink, haste
var FleetfootDancer = newFleetfootDancer

func newFleetfootDancer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Fleetfoot Dancer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.G,
				cost.W,
			}),
			Colors:    []color.Color{color.Green, color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Druid},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.LifelinkStaticBody,
				game.HasteStaticBody,
			},
			OracleText: `
			Trample, lifelink, haste
		`,
		},
	}
}
