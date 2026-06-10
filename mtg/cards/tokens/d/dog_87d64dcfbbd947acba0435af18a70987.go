package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Dog
//
// Type: Token Creature — Dog
//
// Oracle text:
//   Vigilance

// DogToken87d64dcfbbd947acba0435af18a70987 is the card definition for Dog.
var DogToken87d64dcfbbd947acba0435af18a70987 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Dog",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dog},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.VigilanceStaticBody,
		},
		OracleText: `
			Vigilance
		`,
	},
}
