package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MindlessOne is the card definition for Mindless-One.
//
// Type: Token Creature — Missing
//
// Oracle text:

// MindlessOneToken1a0d84f7d7fd4794b6ca565fc111d701 is the card definition for Mindless-One.
var MindlessOneToken1a0d84f7d7fd4794b6ca565fc111d701 = newMindlessOneToken1a0d84f7d7fd4794b6ca565fc111d701()

func newMindlessOneToken1a0d84f7d7fd4794b6ca565fc111d701() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:      "Mindless-One",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sub("Missing")},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
