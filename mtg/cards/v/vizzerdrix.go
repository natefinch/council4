package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Vizzerdrix is the card definition for Vizzerdrix.
//
// Type: Creature — Rabbit Beast
// Cost: {6}{U}
//
// Oracle text:
var Vizzerdrix = newVizzerdrix

func newVizzerdrix() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Vizzerdrix",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rabbit, types.Beast},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
		},
	}
}
