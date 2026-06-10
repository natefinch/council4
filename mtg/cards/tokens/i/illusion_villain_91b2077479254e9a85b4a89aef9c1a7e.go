package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Illusion Villain
//
// Type: Token Creature — Illusion Villain
//
// Oracle text:

// IllusionVillainToken91b2077479254e9a85b4a89aef9c1a7e is the card definition for Illusion Villain.
var IllusionVillainToken91b2077479254e9a85b4a89aef9c1a7e = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Illusion Villain",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Illusion, types.Villain},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	},
}
