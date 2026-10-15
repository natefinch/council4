package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PhyrexianScuta is the card definition for Phyrexian Scuta.
//
// Type: Creature — Phyrexian Zombie
// Cost: {3}{B}
//
// Oracle text:
//
//	Kicker—Pay 3 life. (You may pay 3 life in addition to any other costs as you cast this spell.)
//	If this creature was kicked, it enters with two +1/+1 counters on it.
var PhyrexianScuta = newPhyrexianScuta

func newPhyrexianScuta() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Phyrexian Scuta",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Zombie},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Kicker ",
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{
							AdditionalCosts: []cost.Additional{
								cost.Additional{
									Kind:        cost.AdditionalPayLife,
									Text:        "Pay 3 life",
									Amount:      3,
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("If this creature was kicked, it enters with two +1/+1 counters on it.", &game.Condition{
					EventPermanentWasKicked: true,
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 2}),
			},
			OracleText: `
			Kicker—Pay 3 life. (You may pay 3 life in addition to any other costs as you cast this spell.)
			If this creature was kicked, it enters with two +1/+1 counters on it.
		`,
		},
	}
}
