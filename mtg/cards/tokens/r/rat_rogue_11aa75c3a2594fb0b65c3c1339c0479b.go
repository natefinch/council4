package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Rat Rogue
//
// Type: Token Creature — Rat Rogue
//
// Oracle text:

// RatRogueToken11aa75c3a2594fb0b65c3c1339c0479b is the card definition for Rat Rogue.
var RatRogueToken11aa75c3a2594fb0b65c3c1339c0479b = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Rat Rogue",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Rat, types.Rogue},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
