package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Phytohydra is the card definition for Phytohydra.
//
// Type: Creature — Plant Hydra
// Cost: {2}{G}{W}{W}
//
// Oracle text:
//
//	If damage would be dealt to this creature, put that many +1/+1 counters on it instead.
var Phytohydra = newPhytohydra()

func newPhytohydra() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Phytohydra",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Plant, types.Hydra},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamagePreventionToPlusOneCountersReplacement("If damage would be dealt to this creature, put that many +1/+1 counters on it instead.", false, opt.V[game.Condition]{}),
			},
			OracleText: `
			If damage would be dealt to this creature, put that many +1/+1 counters on it instead.
		`,
		},
	}
}
