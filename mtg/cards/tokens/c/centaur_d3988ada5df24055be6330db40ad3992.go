package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Centaur
//
// Type: Token Enchantment Creature — Centaur
//
// Oracle text:

// CentaurTokend3988ada5df24055be6330db40ad3992 is the card definition for Centaur.
var CentaurTokend3988ada5df24055be6330db40ad3992 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Centaur",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Enchantment, types.Creature},
		Subtypes:  []types.Sub{types.Centaur},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	},
}
