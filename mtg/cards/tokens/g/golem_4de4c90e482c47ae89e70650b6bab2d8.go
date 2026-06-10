package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Golem
//
// Type: Token Artifact Creature — Golem
//
// Oracle text:

// GolemToken4de4c90e482c47ae89e70650b6bab2d8 is the card definition for Golem.
var GolemToken4de4c90e482c47ae89e70650b6bab2d8 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Red),
	CardFace: game.CardFace{
		Name:      "Golem",
		Colors:    []color.Color{color.Red, color.White},
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Golem},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
	},
}
