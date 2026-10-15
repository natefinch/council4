package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IsamaruHoundOfKonda is the card definition for Isamaru, Hound of Konda.
//
// Type: Legendary Creature — Dog
// Cost: {W}
//
// Oracle text:
var IsamaruHoundOfKonda = newIsamaruHoundOfKonda

func newIsamaruHoundOfKonda() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Isamaru, Hound of Konda",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Dog},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
		},
	}
}
