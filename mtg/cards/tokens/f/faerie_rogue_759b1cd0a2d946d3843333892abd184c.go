package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Faerie Rogue
//
// Type: Token Creature — Faerie Rogue
//
// Oracle text:
//   Flying

// FaerieRogueToken759b1cd0a2d946d3843333892abd184c is the card definition for Faerie Rogue.
var FaerieRogueToken759b1cd0a2d946d3843333892abd184c = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue, color.Black),
	CardFace: game.CardFace{
		Name:      "Faerie Rogue",
		Colors:    []color.Color{color.Black, color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Faerie, types.Rogue},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
