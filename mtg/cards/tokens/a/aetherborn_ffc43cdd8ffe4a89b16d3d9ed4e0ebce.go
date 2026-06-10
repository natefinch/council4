package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Aetherborn
//
// Type: Token Creature — Aetherborn
//
// Oracle text:

// AetherbornTokenffc43cdd8ffe4a89b16d3d9ed4e0ebce is the card definition for Aetherborn.
var AetherbornTokenffc43cdd8ffe4a89b16d3d9ed4e0ebce = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Aetherborn",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Aetherborn},
		Power:     opt.Val(game.PT{IsStar: true}),
		Toughness: opt.Val(game.PT{IsStar: true}),
	},
}
