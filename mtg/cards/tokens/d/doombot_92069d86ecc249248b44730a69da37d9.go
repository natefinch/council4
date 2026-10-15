package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Doombot is the card definition for Doombot.
//
// Type: Token Artifact Creature — Robot Villain
//
// Oracle text:

// DoombotToken92069d86ecc249248b44730a69da37d9 is the card definition for Doombot.
var DoombotToken92069d86ecc249248b44730a69da37d9 = newDoombotToken92069d86ecc249248b44730a69da37d9()

func newDoombotToken92069d86ecc249248b44730a69da37d9() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Doombot",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Robot, types.Villain},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
		},
	}
}
