package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BullCerodon is the card definition for Bull Cerodon.
//
// Type: Creature — Beast
// Cost: {4}{R}{W}
//
// Oracle text:
//
//	Vigilance, haste
var BullCerodon = newBullCerodon

func newBullCerodon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Bull Cerodon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.W,
			}),
			Colors:    []color.Color{color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.HasteStaticBody,
			},
			OracleText: `
			Vigilance, haste
		`,
		},
	}
}
