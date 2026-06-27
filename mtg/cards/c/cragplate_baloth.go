package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CragplateBaloth is the card definition for Cragplate Baloth.
//
// Type: Creature — Beast
// Cost: {5}{G}{G}
//
// Oracle text:
//
//	Kicker {2}{G}
//	This spell can't be countered.
//	Hexproof, haste
//	If this creature was kicked, it enters with four +1/+1 counters on it.
var CragplateBaloth = newCragplateBaloth()

func newCragplateBaloth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Cragplate Baloth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(2), cost.G}},
					},
				},
				game.CantBeCounteredStaticBody,
				game.HexproofStaticBody,
				game.HasteStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("If this creature was kicked, it enters with four +1/+1 counters on it.", &game.Condition{
					EventPermanentWasKicked: true,
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 4}),
			},
			OracleText: `
			Kicker {2}{G}
			This spell can't be countered.
			Hexproof, haste
			If this creature was kicked, it enters with four +1/+1 counters on it.
		`,
		},
	}
}
