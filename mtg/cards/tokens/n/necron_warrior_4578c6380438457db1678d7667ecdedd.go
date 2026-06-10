package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Necron Warrior
//
// Type: Token Artifact Creature — Necron Warrior
//
// Oracle text:

// NecronWarriorToken4578c6380438457db1678d7667ecdedd is the card definition for Necron Warrior.
var NecronWarriorToken4578c6380438457db1678d7667ecdedd = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Necron Warrior",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Necron, types.Warrior},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
