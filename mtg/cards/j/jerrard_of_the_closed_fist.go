package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JerrardOfTheClosedFist is the card definition for Jerrard of the Closed Fist.
//
// Type: Legendary Creature — Human Knight
// Cost: {3}{R}{G}{G}
//
// Oracle text:
var JerrardOfTheClosedFist = newJerrardOfTheClosedFist

func newJerrardOfTheClosedFist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Jerrard of the Closed Fist",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Knight},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 5}),
		},
	}
}
