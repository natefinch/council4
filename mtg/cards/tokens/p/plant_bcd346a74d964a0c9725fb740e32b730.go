package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Plant
//
// Type: Token Creature — Plant
//
// Oracle text:

// PlantTokenbcd346a74d964a0c9725fb740e32b730 is the card definition for Plant.
var PlantTokenbcd346a74d964a0c9725fb740e32b730 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Plant",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Plant},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
