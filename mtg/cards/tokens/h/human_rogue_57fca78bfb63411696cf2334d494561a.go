package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Human Rogue
//
// Type: Token Creature — Human Rogue
//
// Oracle text:

// HumanRogueToken57fca78bfb63411696cf2334d494561a is the card definition for Human Rogue.
var HumanRogueToken57fca78bfb63411696cf2334d494561a = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Human Rogue",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Human, types.Rogue},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
