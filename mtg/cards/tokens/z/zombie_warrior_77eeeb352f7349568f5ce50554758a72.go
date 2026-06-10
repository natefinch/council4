package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Zombie Warrior
//
// Type: Token Creature — Zombie Warrior
//
// Oracle text:
//   Vigilance

// ZombieWarriorToken77eeeb352f7349568f5ce50554758a72 is the card definition for Zombie Warrior.
var ZombieWarriorToken77eeeb352f7349568f5ce50554758a72 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Zombie Warrior",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Zombie, types.Warrior},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.VigilanceStaticBody,
		},
		OracleText: `
			Vigilance
		`,
	},
}
