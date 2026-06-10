package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Ragavan
//
// Type: Token Legendary Creature — Monkey
//
// Oracle text:

// RagavanTokene2a98673a9c24579b962080405bbb661 is the card definition for Ragavan.
var RagavanTokene2a98673a9c24579b962080405bbb661 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:       "Ragavan",
		Colors:     []color.Color{color.Red},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Monkey},
		Power:      opt.Val(game.PT{Value: 2}),
		Toughness:  opt.Val(game.PT{Value: 1}),
	},
}
