package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Sliver Army
//
// Type: Token Creature — Sliver Army
//
// Oracle text:

// SliverArmyToken5118bcdd9f894a19b18a2257b3b0be0d is the card definition for Sliver Army.
var SliverArmyToken5118bcdd9f894a19b18a2257b3b0be0d = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Sliver Army",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Sliver, types.Army},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 0}),
	},
}
