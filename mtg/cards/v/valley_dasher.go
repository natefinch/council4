package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ValleyDasher is the card definition for Valley Dasher.
//
// Type: Creature — Human Berserker
// Cost: {1}{R}
//
// Oracle text:
//
//	Haste
//	This creature attacks each combat if able.
var ValleyDasher = newValleyDasher

func newValleyDasher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Valley Dasher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Berserker},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
				game.MustAttackStaticBody,
			},
			OracleText: `
			Haste
			This creature attacks each combat if able.
		`,
		},
	}
}
