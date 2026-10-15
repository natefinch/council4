package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SivitriScarzam is the card definition for Sivitri Scarzam.
//
// Type: Legendary Creature — Human
// Cost: {5}{U}{B}
//
// Oracle text:
var SivitriScarzam = newSivitriScarzam

func newSivitriScarzam() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Sivitri Scarzam",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
				cost.B,
			}),
			Colors:     []color.Color{color.Black, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 4}),
		},
	}
}
