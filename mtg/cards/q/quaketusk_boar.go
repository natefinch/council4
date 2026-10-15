package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// QuaketuskBoar is the card definition for Quaketusk Boar.
//
// Type: Creature — Elemental Boar
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Reach, trample, haste
var QuaketuskBoar = newQuaketuskBoar

func newQuaketuskBoar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Quaketusk Boar",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental, types.Boar},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.TrampleStaticBody,
				game.HasteStaticBody,
			},
			OracleText: `
			Reach, trample, haste
		`,
		},
	}
}
