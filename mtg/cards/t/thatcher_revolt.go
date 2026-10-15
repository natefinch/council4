package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThatcherRevolt is the card definition for Thatcher Revolt.
//
// Type: Sorcery
// Cost: {2}{R}
//
// Oracle text:
//
//	Create three 1/1 red Human creature tokens with haste. Sacrifice those tokens at the beginning of the next end step.
var ThatcherRevolt = newThatcherRevolt

func newThatcherRevolt() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Thatcher Revolt",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount:        game.Fixed(3),
							Source:        game.TokenDef(thatcherRevoltToken),
							PublishLinked: game.LinkedKey("sequence-effect-0-product"),
						},
						LocalProducts: game.LocalProducts{
							Links: []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing: game.DelayedAtBeginningOfNextEndStep,
								Content: game.AbilityContent{
									Modes: []game.Mode{
										game.Mode{
											Sequence: []game.Instruction{
												game.Instruction{
													Primitive: game.Sacrifice{
														Group: game.CapturedObjectsGroup(),
													},
												},
											},
										},
									},
									MinModes: 1,
									MaxModes: 1,
								},
								CapturedObjectGroup: opt.Val(game.LinkedObjectReference("sequence-effect-0-product")),
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create three 1/1 red Human creature tokens with haste. Sacrifice those tokens at the beginning of the next end step.
		`,
		},
	}
}

var thatcherRevoltToken = newThatcherRevoltToken()

func newThatcherRevoltToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Human",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
		},
	}
}
