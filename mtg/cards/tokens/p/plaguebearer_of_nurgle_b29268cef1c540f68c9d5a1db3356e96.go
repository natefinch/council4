package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Plaguebearer of Nurgle
//
// Type: Token Creature — Demon
//
// Oracle text:

// PlaguebearerOfNurgleTokenb29268cef1c540f68c9d5a1db3356e96 is the card definition for Plaguebearer of Nurgle.
var PlaguebearerOfNurgleTokenb29268cef1c540f68c9d5a1db3356e96 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Plaguebearer of Nurgle",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Demon},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 3}),
	},
}
