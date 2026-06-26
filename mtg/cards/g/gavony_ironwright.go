package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GavonyIronwright is the card definition for Gavony Ironwright.
//
// Type: Creature — Human Soldier
// Cost: {2}{W}
//
// Oracle text:
//
//	Fateful hour — As long as you have 5 or less life, other creatures you control get +1/+4.
var GavonyIronwright = newGavonyIronwright()

func newGavonyIronwright() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Gavony Ironwright",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerLife, Op: compare.LessOrEqual, Value: 5}},
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 4,
						},
					},
				},
			},
			OracleText: `
			Fateful hour — As long as you have 5 or less life, other creatures you control get +1/+4.
		`,
		},
	}
}
