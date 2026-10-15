package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BarktoothWarbeard is the card definition for Barktooth Warbeard.
//
// Type: Legendary Creature — Human Warrior
// Cost: {4}{B}{R}{R}
//
// Oracle text:
var BarktoothWarbeard = newBarktoothWarbeard

func newBarktoothWarbeard() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Barktooth Warbeard",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.R,
				cost.R,
			}),
			Colors:     []color.Color{color.Black, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Warrior},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 5}),
		},
	}
}
