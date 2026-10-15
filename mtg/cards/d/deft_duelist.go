package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DeftDuelist is the card definition for Deft Duelist.
//
// Type: Creature — Human Rogue
// Cost: {W}{U}
//
// Oracle text:
//
//	First strike
//	Shroud (This creature can't be the target of spells or abilities.)
var DeftDuelist = newDeftDuelist

func newDeftDuelist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Deft Duelist",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Rogue},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.ShroudStaticBody,
			},
			OracleText: `
			First strike
			Shroud (This creature can't be the target of spells or abilities.)
		`,
		},
	}
}
