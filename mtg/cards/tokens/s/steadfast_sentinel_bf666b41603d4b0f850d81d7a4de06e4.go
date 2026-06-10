package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Steadfast Sentinel
//
// Type: Token Creature — Zombie Human Cleric
//
// Oracle text:
//   Vigilance

// SteadfastSentinelTokenbf666b41603d4b0f850d81d7a4de06e4 is the card definition for Steadfast Sentinel.
var SteadfastSentinelTokenbf666b41603d4b0f850d81d7a4de06e4 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Steadfast Sentinel",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Zombie, types.Human, types.Cleric},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.VigilanceStaticBody,
		},
		OracleText: `
			Vigilance
		`,
	},
}
