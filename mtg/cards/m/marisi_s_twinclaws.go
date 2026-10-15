package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MarisiSTwinclaws is the card definition for Marisi's Twinclaws.
//
// Type: Creature — Cat Warrior
// Cost: {2}{R/W}{G}
//
// Oracle text:
//
//	Double strike
var MarisiSTwinclaws = newMarisiSTwinclaws

func newMarisiSTwinclaws() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Marisi's Twinclaws",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.R, mana.W),
				cost.G,
			}),
			Colors:    []color.Color{color.Green, color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
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
