package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NacatlWarPride is the card definition for Nacatl War-Pride.
//
// Type: Creature — Cat Warrior
// Cost: {3}{G}{G}{G}
//
// Oracle text:
//
//	This creature must be blocked by exactly one creature if able.
//	Whenever this creature attacks, create X tokens that are copies of it and that are tapped and attacking, where X is the number of creatures defending player controls. Exile the tokens at the beginning of the next end step.
var NacatlWarPride = newNacatlWarPride

func newNacatlWarPride() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Nacatl War-Pride",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.MustBeBlockedByExactlyOneStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 1,
										Group:      game.PlayerControlledGroup(game.DefendingPlayerReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
									}),
									Source: game.TokenCopyOf(game.TokenCopySpec{
										Source: game.TokenCopySourceObject,
										Object: game.SourcePermanentReference(),
									}),
									EntryTapped:            true,
									EntryAttackingDefender: opt.Val(game.DefendingPlayerReference()),
									PublishLinked:          game.LinkedKey("delayed-created-token-exile-1"),
								},
							},
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										Timing:              game.DelayedAtBeginningOfNextEndStep,
										CapturedObjectGroup: opt.Val(game.LinkedObjectReference("delayed-created-token-exile-1")),
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.Exile{
														Group: game.CapturedObjectsGroup(),
													},
												},
											},
										}.Ability(),
									},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			This creature must be blocked by exactly one creature if able.
			Whenever this creature attacks, create X tokens that are copies of it and that are tapped and attacking, where X is the number of creatures defending player controls. Exile the tokens at the beginning of the next end step.
		`,
		},
	}
}
