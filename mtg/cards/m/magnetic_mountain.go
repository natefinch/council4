package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MagneticMountain is the card definition for Magnetic Mountain.
//
// Type: Enchantment
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	Blue creatures don't untap during their controllers' untap steps.
//	At the beginning of each player's upkeep, that player may choose any number of tapped blue creatures they control and pay {4} for each creature chosen this way. If the player does, untap those creatures.
var MagneticMountain = newMagneticMountain

func newMagneticMountain() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Magnetic Mountain",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Blue creatures don't untap during their controllers' untap steps.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:              game.RuleEffectDoesntUntap,
							Duration:          game.DurationPermanent,
							PermanentTypes:    []types.Card{types.Creature},
							AffectedSelection: game.Selection{ColorsAny: []color.Color{color.Blue}},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of each player's upkeep, that player may choose any number of tapped blue creatures they control and pay {4} for each creature chosen this way. If the player does, untap those creatures.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepUpkeep,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.PayRepeatedly{
											Payment: game.ResolutionPayment{
												Prompt: "Pay {4}?",
												Payer:  opt.Val(game.EventPlayerReference()),
												ManaCost: opt.Val(cost.Mana{
													cost.O(4),
												}),
											},
											PublishCount: game.ResultKey("per-creature-untap-count"),
											Prompt:       "Pay {4}?",
										},
										PublishResult: game.ResultKey("per-creature-untap-count"),
									},
									game.Instruction{
										Primitive: game.Untap{
											Group:      game.PlayerControlledGroup(game.EventPlayerReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.Blue}, Tapped: game.TriTrue}),
											ChooseUpTo: true,
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:      game.DynamicAmountChosenNumber,
												ResultKey: game.ResultKey("per-creature-untap-count"),
											}),
											Chooser: game.EventPlayerReference(),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("per-creature-untap-count"),
											Succeeded: game.TriTrue,
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
			Blue creatures don't untap during their controllers' untap steps.
			At the beginning of each player's upkeep, that player may choose any number of tapped blue creatures they control and pay {4} for each creature chosen this way. If the player does, untap those creatures.
		`,
		},
	}
}
