package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Phyrexian Goblin
//
// Type: Token Creature — Phyrexian Goblin
//
// Oracle text:

// PhyrexianGoblinTokenb43340e3bb7645018db2040b7cae33b7 is the card definition for Phyrexian Goblin.
var PhyrexianGoblinTokenb43340e3bb7645018db2040b7cae33b7 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Phyrexian Goblin",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Phyrexian, types.Goblin},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
