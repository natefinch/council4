package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GnarlidColony is the card definition for Gnarlid Colony.
//
// Type: Creature — Beast
// Cost: {1}{G}
//
// Oracle text:
//
//	Kicker {2}{G} (You may pay an additional {2}{G} as you cast this spell.)
//	If this creature was kicked, it enters with two +1/+1 counters on it.
//	Each creature you control with a +1/+1 counter on it has trample. (It can deal excess combat damage to the player or planeswalker it's attacking.)
var GnarlidColony = newGnarlidColony()

func newGnarlidColony() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Gnarlid Colony",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(2), cost.G}},
					},
				},
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchCounter: true, RequiredCounter: counter.PlusOnePlusOne}),
							AddKeywords: []game.Keyword{
								game.Trample,
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
			Kicker {2}{G} (You may pay an additional {2}{G} as you cast this spell.)
			If this creature was kicked, it enters with two +1/+1 counters on it.
			Each creature you control with a +1/+1 counter on it has trample. (It can deal excess combat damage to the player or planeswalker it's attacking.)
		`,
		},
	}
}
