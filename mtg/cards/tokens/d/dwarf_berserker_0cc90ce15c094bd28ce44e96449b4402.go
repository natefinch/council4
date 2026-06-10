package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Dwarf Berserker
//
// Type: Token Creature — Dwarf Berserker
//
// Oracle text:

// DwarfBerserkerToken0cc90ce15c094bd28ce44e96449b4402 is the card definition for Dwarf Berserker.
var DwarfBerserkerToken0cc90ce15c094bd28ce44e96449b4402 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Dwarf Berserker",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dwarf, types.Berserker},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
