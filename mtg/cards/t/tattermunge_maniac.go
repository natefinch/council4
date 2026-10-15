package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TattermungeManiac is the card definition for Tattermunge Maniac.
//
// Type: Creature — Goblin Warrior
// Cost: {R/G}
//
// Oracle text:
//
//	This creature attacks each combat if able.
var TattermungeManiac = newTattermungeManiac

func newTattermungeManiac() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Tattermunge Maniac",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.R, mana.G),
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.MustAttackStaticBody,
			},
			OracleText: `
			This creature attacks each combat if able.
		`,
		},
	}
}
