package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Gigantosaurus is the card definition for Gigantosaurus.
//
// Type: Creature — Dinosaur
// Cost: {G}{G}{G}{G}{G}
//
// Oracle text:
var Gigantosaurus = newGigantosaurus

func newGigantosaurus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Gigantosaurus",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.G,
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dinosaur},
			Power:     opt.Val(game.PT{Value: 10}),
			Toughness: opt.Val(game.PT{Value: 10}),
		},
	}
}
