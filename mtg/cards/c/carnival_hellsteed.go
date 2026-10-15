package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CarnivalHellsteed is the card definition for Carnival Hellsteed.
//
// Type: Creature — Nightmare Horse
// Cost: {4}{B}{R}
//
// Oracle text:
//
//	First strike, haste
//	Unleash (You may have this creature enter with a +1/+1 counter on it. It can't block as long as it has a +1/+1 counter on it.)
var CarnivalHellsteed = newCarnivalHellsteed

func newCarnivalHellsteed() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Carnival Hellsteed",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.R,
			}),
			Colors:    []color.Color{color.Black, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Nightmare, types.Horse},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.HasteStaticBody,
				game.UnleashStaticBody,
			},
			OracleText: `
			First strike, haste
			Unleash (You may have this creature enter with a +1/+1 counter on it. It can't block as long as it has a +1/+1 counter on it.)
		`,
		},
	}
}
