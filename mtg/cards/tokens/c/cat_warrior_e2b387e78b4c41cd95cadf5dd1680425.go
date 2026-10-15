package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CatWarrior is the card definition for Cat Warrior.
//
// Type: Token Creature — Cat Warrior
//
// Oracle text:
//   Forestwalk

// CatWarriorTokene2b387e78b4c41cd95cadf5dd1680425 is the card definition for Cat Warrior.
var CatWarriorTokene2b387e78b4c41cd95cadf5dd1680425 = newCatWarriorTokene2b387e78b4c41cd95cadf5dd1680425()

func newCatWarriorTokene2b387e78b4c41cd95cadf5dd1680425() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:      "Cat Warrior",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ForestwalkStaticBody,
			},
			OracleText: `
			Forestwalk
		`,
		},
	}
}
