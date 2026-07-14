package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Demoralize is the card definition for Demoralize.
//
// Type: Instant
// Cost: {2}{R}
//
// Oracle text:
//
//	All creatures gain menace until end of turn. (They can't be blocked except by two or more creatures.)
//	Threshold — If there are seven or more cards in your graveyard, creatures can't block this turn.
var Demoralize = newDemoralize

func newDemoralize() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Demoralize",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
									AddKeywords: []game.Keyword{
										game.Menace,
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.ApplyRule{
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind:           game.RuleEffectCantBlock,
									PermanentTypes: []types.Card{types.Creature},
								},
							},
							Duration: game.DurationThisTurn,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerGraveyardCardCount, Op: compare.GreaterOrEqual, Value: 7}},
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			All creatures gain menace until end of turn. (They can't be blocked except by two or more creatures.)
			Threshold — If there are seven or more cards in your graveyard, creatures can't block this turn.
		`,
		},
	}
}
