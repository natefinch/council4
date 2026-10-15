package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RootMaze is the card definition for Root Maze.
//
// Type: Enchantment
// Cost: {G}
//
// Oracle text:
//
//	Artifacts and lands enter tapped.
var RootMaze = newRootMaze

func newRootMaze() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Root Maze",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedGroupReplacement("Artifacts and lands enter tapped.", game.TriggerControllerAny, types.Artifact, types.Land),
			},
			OracleText: `
			Artifacts and lands enter tapped.
		`,
		},
	}
}
