package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Ogre Warrior
//
// Type: Token Creature — Ogre Warrior
//
// Oracle text:

// OgreWarriorToken77fc8f50dcf14d608fd6b1c7d58a9724 is the card definition for Ogre Warrior.
var OgreWarriorToken77fc8f50dcf14d608fd6b1c7d58a9724 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Ogre Warrior",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Ogre, types.Warrior},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 3}),
	},
}
