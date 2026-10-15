package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Giant is the card definition for Giant.
//
// Type: Token Creature — Giant
//
// Oracle text:

// GiantToken22be49fa33a647fa9bfcf06e044d3cdc is the card definition for Giant.
var GiantToken22be49fa33a647fa9bfcf06e044d3cdc = newGiantToken22be49fa33a647fa9bfcf06e044d3cdc()

func newGiantToken22be49fa33a647fa9bfcf06e044d3cdc() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name:      "Giant",
			Colors:    []color.Color{color.Red, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Giant},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
		},
	}
}
