package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AutochthonWurm is the card definition for Autochthon Wurm.
//
// Type: Creature — Wurm
// Cost: {10}{G}{G}{G}{W}{W}
//
// Oracle text:
//
//	Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
//	Trample
var AutochthonWurm = newAutochthonWurm

func newAutochthonWurm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Autochthon Wurm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(10),
				cost.G,
				cost.G,
				cost.G,
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wurm},
			Power:     opt.Val(game.PT{Value: 9}),
			Toughness: opt.Val(game.PT{Value: 14}),
			StaticAbilities: []game.StaticAbility{
				game.ConvokeStaticBody,
				game.TrampleStaticBody,
			},
			OracleText: `
			Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
			Trample
		`,
		},
	}
}
