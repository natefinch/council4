package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FeedTheClan is the card definition for Feed the Clan.
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	You gain 5 life.
//	Ferocious — You gain 10 life instead if you control a creature with power 4 or greater.
var FeedTheClan = newFeedTheClan

func newFeedTheClan() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Feed the Clan",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(5),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate: true,
								ControlsMatching: opt.Val(game.SelectionCount{
									Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})},
								}),
							}),
						}),
						PublishCondition: "condition-0",
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(10),
							Player: game.ControllerReference(),
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
				},
			}.Ability()),
			OracleText: `
			You gain 5 life.
			Ferocious — You gain 10 life instead if you control a creature with power 4 or greater.
		`,
		},
	}
}
