package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TerraStomper is the card definition for Terra Stomper.
//
// Type: Creature — Beast
// Cost: {3}{G}{G}{G}
//
// Oracle text:
//
//	This spell can't be countered.
//	Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
var TerraStomper = newTerraStomper

func newTerraStomper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Terra Stomper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 8}),
			Toughness: opt.Val(game.PT{Value: 8}),
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
				game.TrampleStaticBody,
			},
			OracleText: `
			This spell can't be countered.
			Trample (This creature can deal excess combat damage to the player or planeswalker it's attacking.)
		`,
		},
	}
}
