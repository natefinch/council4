package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Kithkin Soldier
//
// Type: Token Creature — Kithkin Soldier
//
// Oracle text:

// KithkinSoldierTokencaf4a0d264de42dfa88ebb0d3f34e207 is the card definition for Kithkin Soldier.
var KithkinSoldierTokencaf4a0d264de42dfa88ebb0d3f34e207 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Kithkin Soldier",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Kithkin, types.Soldier},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
