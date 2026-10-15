package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LadyOrca is the card definition for Lady Orca.
//
// Type: Legendary Creature — Demon
// Cost: {5}{B}{R}
//
// Oracle text:
var LadyOrca = newLadyOrca

func newLadyOrca() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Lady Orca",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.B,
				cost.R,
			}),
			Colors:     []color.Color{color.Black, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Demon},
			Power:      opt.Val(game.PT{Value: 7}),
			Toughness:  opt.Val(game.PT{Value: 4}),
		},
	}
}
