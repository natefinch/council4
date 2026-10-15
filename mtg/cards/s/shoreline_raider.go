package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShorelineRaider is the card definition for Shoreline Raider.
//
// Type: Creature — Merfolk
// Cost: {2}{U}
//
// Oracle text:
//
//	Protection from Kavu
var ShorelineRaider = newShorelineRaider

func newShorelineRaider() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Shoreline Raider",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ProtectionFromSubtypesStaticAbility(types.Kavu),
			},
			OracleText: `
			Protection from Kavu
		`,
		},
	}
}
