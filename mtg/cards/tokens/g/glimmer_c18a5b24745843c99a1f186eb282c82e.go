package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Glimmer
//
// Type: Token Enchantment Creature — Glimmer
//
// Oracle text:

// GlimmerTokenc18a5b24745843c99a1f186eb282c82e is the card definition for Glimmer.
var GlimmerTokenc18a5b24745843c99a1f186eb282c82e = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Glimmer",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Enchantment, types.Creature},
		Subtypes:  []types.Sub{types.Glimmer},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
