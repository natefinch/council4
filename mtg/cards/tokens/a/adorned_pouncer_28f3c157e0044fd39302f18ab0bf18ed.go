package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AdornedPouncer is the card definition for Adorned Pouncer.
//
// Type: Token Creature — Zombie Cat
//
// Oracle text:
//   Double strike

// AdornedPouncerToken28f3c157e0044fd39302f18ab0bf18ed is the card definition for Adorned Pouncer.
var AdornedPouncerToken28f3c157e0044fd39302f18ab0bf18ed = newAdornedPouncerToken28f3c157e0044fd39302f18ab0bf18ed()

func newAdornedPouncerToken28f3c157e0044fd39302f18ab0bf18ed() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:      "Adorned Pouncer",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Cat},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.DoubleStrikeStaticBody,
			},
			OracleText: `
			Double strike
		`,
		},
	}
}
