package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Fish
//
// Type: Token Creature — Fish
//
// Oracle text:
//   This creature can't be blocked.

// FishToken5bef2c1c72154b018bb3426e1c3dd2e0 is the card definition for Fish.
var FishToken5bef2c1c72154b018bb3426e1c3dd2e0 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Fish",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Fish},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.CantBeBlockedStaticBody,
		},
		OracleText: `
			This creature can't be blocked.
		`,
	},
}
