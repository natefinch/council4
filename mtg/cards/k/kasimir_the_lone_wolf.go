package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KasimirTheLoneWolf is the card definition for Kasimir the Lone Wolf.
//
// Type: Legendary Creature — Human Warrior
// Cost: {4}{W}{U}
//
// Oracle text:
var KasimirTheLoneWolf = newKasimirTheLoneWolf

func newKasimirTheLoneWolf() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Kasimir the Lone Wolf",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Warrior},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 3}),
		},
	}
}
