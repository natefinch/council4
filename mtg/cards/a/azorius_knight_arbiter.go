package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AzoriusKnightArbiter is the card definition for Azorius Knight-Arbiter.
//
// Type: Creature — Human Knight
// Cost: {3}{W}{U}
//
// Oracle text:
//
//	Vigilance
//	This creature can't be blocked.
var AzoriusKnightArbiter = newAzoriusKnightArbiter

func newAzoriusKnightArbiter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Azorius Knight-Arbiter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.CantBeBlockedStaticBody,
			},
			OracleText: `
			Vigilance
			This creature can't be blocked.
		`,
		},
	}
}
