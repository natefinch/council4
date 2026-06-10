package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Demon
//
// Type: Token Creature — Demon
//
// Oracle text:
//   Flying

// DemonToken20f3d81fe0cc472cbbd24a2f380a5dee is the card definition for Demon.
var DemonToken20f3d81fe0cc472cbbd24a2f380a5dee = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Demon",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Demon},
		Power:     opt.Val(game.PT{Value: 6}),
		Toughness: opt.Val(game.PT{Value: 6}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
