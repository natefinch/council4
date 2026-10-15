package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CarnivorousPlant is the card definition for Carnivorous Plant.
//
// Type: Creature — Plant Wall
// Cost: {3}{G}
//
// Oracle text:
//
//	Defender
var CarnivorousPlant = newCarnivorousPlant

func newCarnivorousPlant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Carnivorous Plant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Plant, types.Wall},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			OracleText: `
			Defender
		`,
		},
	}
}
