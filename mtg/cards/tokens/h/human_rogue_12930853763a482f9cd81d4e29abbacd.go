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

// HumanRogueToken12930853763a482f9cd81d4e29abbacd is the card definition for Human Rogue.
var HumanRogueToken12930853763a482f9cd81d4e29abbacd = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Human Rogue",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Human, types.Rogue},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
