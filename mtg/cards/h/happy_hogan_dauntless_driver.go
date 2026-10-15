package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HappyHoganDauntlessDriver is the card definition for Happy Hogan, Dauntless Driver.
//
// Type: Legendary Creature — Human Pilot
// Cost: {R}
//
// Oracle text:
var HappyHoganDauntlessDriver = newHappyHoganDauntlessDriver

func newHappyHoganDauntlessDriver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Happy Hogan, Dauntless Driver",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Pilot},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 1}),
		},
	}
}
