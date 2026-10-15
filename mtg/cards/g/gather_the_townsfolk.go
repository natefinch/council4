package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GatherTheTownsfolk is the card definition for Gather the Townsfolk.
//
// Type: Sorcery
// Cost: {1}{W}
//
// Oracle text:
//
//	Create two 1/1 white Human creature tokens.
//	Fateful hour — If you have 5 or less life, create five of those tokens instead.
var GatherTheTownsfolk = newGatherTheTownsfolk

func newGatherTheTownsfolk() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Gather the Townsfolk",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(gatherTheTownsfolkToken),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:     true,
								Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerLife, Op: compare.LessOrEqual, Value: 5}},
							}),
						}),
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(5),
							Source: game.TokenDef(gatherTheTownsfolkToken),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerLife, Op: compare.LessOrEqual, Value: 5}},
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Create two 1/1 white Human creature tokens.
			Fateful hour — If you have 5 or less life, create five of those tokens instead.
		`,
		},
	}
}

var gatherTheTownsfolkToken = newGatherTheTownsfolkToken()

func newGatherTheTownsfolkToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Human",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
