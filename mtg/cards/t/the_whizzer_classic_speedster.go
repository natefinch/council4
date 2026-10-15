package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TheWhizzerClassicSpeedster is the card definition for The Whizzer, Classic Speedster.
//
// Type: Legendary Creature — Human Hero
// Cost: {3}{R}
//
// Oracle text:
//
//	First strike, haste
var TheWhizzerClassicSpeedster = newTheWhizzerClassicSpeedster

func newTheWhizzerClassicSpeedster() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "The Whizzer, Classic Speedster",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Hero},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.HasteStaticBody,
			},
			OracleText: `
			First strike, haste
		`,
		},
	}
}
