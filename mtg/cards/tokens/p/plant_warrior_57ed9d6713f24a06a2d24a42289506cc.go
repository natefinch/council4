package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Plant Warrior
//
// Type: Token Creature — Plant Warrior
//
// Oracle text:
//   Reach

// PlantWarriorToken57ed9d6713f24a06a2d24a42289506cc is the card definition for Plant Warrior.
var PlantWarriorToken57ed9d6713f24a06a2d24a42289506cc = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Plant Warrior",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Plant, types.Warrior},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.ReachStaticBody,
		},
		OracleText: `
			Reach
		`,
	},
}
