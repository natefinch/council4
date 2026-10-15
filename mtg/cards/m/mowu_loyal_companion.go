package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MowuLoyalCompanion is the card definition for Mowu, Loyal Companion.
//
// Type: Legendary Creature — Dog
// Cost: {3}{G}
//
// Oracle text:
//
//	Vigilance, trample
//	If one or more +1/+1 counters would be put on Mowu, that many plus one +1/+1 counters are put on it instead.
var MowuLoyalCompanion = newMowuLoyalCompanion

func newMowuLoyalCompanion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Mowu, Loyal Companion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Dog},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.TrampleStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.SelfCounterPlacementReplacement("If one or more +1/+1 counters would be put on Mowu, that many plus one +1/+1 counters are put on it instead.", 0, 1, counter.PlusOnePlusOne),
			},
			OracleText: `
			Vigilance, trample
			If one or more +1/+1 counters would be put on Mowu, that many plus one +1/+1 counters are put on it instead.
		`,
		},
	}
}
