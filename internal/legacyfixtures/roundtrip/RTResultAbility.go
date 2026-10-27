// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTResultAbility is the card definition for RT Result Ability.
//
// Type: Enchantment
//
// Oracle text:
//
//	{1}: Discard three cards. If two land cards were discarded this way, draw a card.
var RTResultAbility = newRTResultAbility

func newRTResultAbility() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Result Ability",
			Types: []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}: Discard three cards. If two land cards were discarded this way, draw a card.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Discard{
											Amount: game.Fixed(3),
											Player: game.ControllerReference(),
										},
										PublishResult: game.ResultKey("if-you-do"),
										LocalProducts: game.LocalProducts{
											Results: []game.ResultKey{game.ResultKey("if-you-do")},
										},
									},
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:             game.ResultKey("if-you-do"),
											Succeeded:       game.TriTrue,
											ObjectSelection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}}),
											ObjectCountRange: opt.Val(game.IntRange{
												Min: 2,
												Max: 2,
											}),
											CardOnly: true,
										}),
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			{1}: Discard three cards. If two land cards were discarded this way, draw a card.
		`,
		},
	}
}
