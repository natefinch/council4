package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Tyranid Gargoyle
//
// Type: Token Creature — Tyranid Gargoyle
//
// Oracle text:
//   Flying

// TyranidGargoyleToken5fe924f8880c425c9daed2f5455a5f2c is the card definition for Tyranid Gargoyle.
var TyranidGargoyleToken5fe924f8880c425c9daed2f5455a5f2c = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Tyranid Gargoyle",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Tyranid, types.Gargoyle},
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
