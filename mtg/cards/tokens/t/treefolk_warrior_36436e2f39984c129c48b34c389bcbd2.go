package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TreefolkWarrior is the card definition for Treefolk Warrior.
//
// Type: Token Creature — Treefolk Warrior
//
// Oracle text:
//   This creature's power and toughness are each equal to the number of Forests you control.

// TreefolkWarriorToken36436e2f39984c129c48b34c389bcbd2 is the card definition for Treefolk Warrior.
var TreefolkWarriorToken36436e2f39984c129c48b34c389bcbd2 = newTreefolkWarriorToken36436e2f39984c129c48b34c389bcbd2()

func newTreefolkWarriorToken36436e2f39984c129c48b34c389bcbd2() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:             "Treefolk Warrior",
			Colors:           []color.Color{color.Green},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Treefolk, types.Warrior},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerSubtypeCount, Subtype: types.Forest}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerSubtypeCount, Subtype: types.Forest}),
			OracleText: `
			This creature's power and toughness are each equal to the number of Forests you control.
		`,
		},
	}
}
