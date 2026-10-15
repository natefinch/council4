package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// QuakestriderCeratops is the card definition for Quakestrider Ceratops.
//
// Type: Creature — Dinosaur
// Cost: {3}{G}{G}{G}
//
// Oracle text:
var QuakestriderCeratops = newQuakestriderCeratops

func newQuakestriderCeratops() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Quakestrider Ceratops",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dinosaur},
			Power:     opt.Val(game.PT{Value: 12}),
			Toughness: opt.Val(game.PT{Value: 8}),
		},
	}
}
