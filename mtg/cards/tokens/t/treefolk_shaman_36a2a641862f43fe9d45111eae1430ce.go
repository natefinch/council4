package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Treefolk Shaman
//
// Type: Token Creature — Treefolk Shaman
//
// Oracle text:

// TreefolkShamanToken36a2a641862f43fe9d45111eae1430ce is the card definition for Treefolk Shaman.
var TreefolkShamanToken36a2a641862f43fe9d45111eae1430ce = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Treefolk Shaman",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Treefolk, types.Shaman},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 5}),
	},
}
