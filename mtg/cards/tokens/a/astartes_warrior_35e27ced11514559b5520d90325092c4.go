package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Astartes Warrior
//
// Type: Token Creature — Astartes Warrior
//
// Oracle text:
//   Vigilance

// AstartesWarriorToken35e27ced11514559b5520d90325092c4 is the card definition for Astartes Warrior.
var AstartesWarriorToken35e27ced11514559b5520d90325092c4 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Astartes Warrior",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Astartes, types.Warrior},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.VigilanceStaticBody,
		},
		OracleText: `
			Vigilance
		`,
	},
}
