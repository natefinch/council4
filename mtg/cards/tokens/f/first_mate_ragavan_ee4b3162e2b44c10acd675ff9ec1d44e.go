package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// First Mate Ragavan
//
// Type: Token Legendary Creature — Monkey Pirate
//
// Oracle text:

// FirstMateRagavanTokenee4b3162e2b44c10acd675ff9ec1d44e is the card definition for First Mate Ragavan.
var FirstMateRagavanTokenee4b3162e2b44c10acd675ff9ec1d44e = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:       "First Mate Ragavan",
		Colors:     []color.Color{color.Red},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Monkey, types.Pirate},
		Power:      opt.Val(game.PT{Value: 2}),
		Toughness:  opt.Val(game.PT{Value: 1}),
	},
}
