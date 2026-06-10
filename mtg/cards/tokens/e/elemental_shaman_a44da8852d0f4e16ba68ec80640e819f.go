package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Elemental Shaman
//
// Type: Token Creature — Elemental Shaman
//
// Oracle text:

// ElementalShamanTokena44da8852d0f4e16ba68ec80640e819f is the card definition for Elemental Shaman.
var ElementalShamanTokena44da8852d0f4e16ba68ec80640e819f = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Elemental Shaman",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elemental, types.Shaman},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
