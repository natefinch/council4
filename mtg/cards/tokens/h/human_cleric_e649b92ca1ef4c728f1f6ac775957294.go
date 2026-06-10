package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Human Cleric
//
// Type: Token Creature — Human Cleric
//
// Oracle text:

// HumanClericTokene649b92ca1ef4c728f1f6ac775957294 is the card definition for Human Cleric.
var HumanClericTokene649b92ca1ef4c728f1f6ac775957294 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Black),
	CardFace: game.CardFace{
		Name:      "Human Cleric",
		Colors:    []color.Color{color.Black, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Human, types.Cleric},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
