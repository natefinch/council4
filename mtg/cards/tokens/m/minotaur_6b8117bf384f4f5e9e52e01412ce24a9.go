package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Minotaur
//
// Type: Token Creature — Minotaur
//
// Oracle text:

// MinotaurToken6b8117bf384f4f5e9e52e01412ce24a9 is the card definition for Minotaur.
var MinotaurToken6b8117bf384f4f5e9e52e01412ce24a9 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Minotaur",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Minotaur},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 3}),
	},
}
