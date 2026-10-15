package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DreamTides is the card definition for Dream Tides.
//
// Type: Enchantment
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	Creatures don't untap during their controllers' untap steps.
//	At the beginning of each player's upkeep, that player may choose any number of tapped nongreen creatures they control and pay {2} for each creature chosen this way. If the player does, untap those creatures.
var DreamTides = newDreamTides

func newDreamTides() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Dream Tides",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures don't untap during their controllers' untap steps.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectDoesntUntap,
							Duration:       game.DurationPermanent,
							PermanentTypes: []types.Card{types.Creature},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of each player's upkeep, that player may choose any number of tapped nongreen creatures they control and pay {2} for each creature chosen this way. If the player does, untap those creatures.",
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
												Prompt: "Pay {2}?",
												Payer:  opt.Val(game.EventPlayerReference()),
												ManaCost: opt.Val(cost.Mana{
													cost.O(2),
												}),
											},
											PublishCount: game.ResultKey("per-creature-untap-count"),
											Prompt:       "Pay {2}?",
										},
										PublishResult: game.ResultKey("per-creature-untap-count"),
									},
									game.Instruction{
										Primitive: game.Untap{
											Group:      game.PlayerControlledGroup(game.EventPlayerReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedColors: []color.Color{color.Green}, Tapped: game.TriTrue}),
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
			Creatures don't untap during their controllers' untap steps.
			At the beginning of each player's upkeep, that player may choose any number of tapped nongreen creatures they control and pay {2} for each creature chosen this way. If the player does, untap those creatures.
		`,
		},
	}
}
