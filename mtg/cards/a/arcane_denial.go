package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArcaneDenial is the card definition for Arcane Denial.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Counter target spell. Its controller may draw up to two cards at the beginning of the next turn's upkeep.
//	You draw a card at the beginning of the next turn's upkeep.
var ArcaneDenial = newArcaneDenial

func newArcaneDenial() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Arcane Denial",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							StackObjectKinds: []game.StackObjectKind{game.StackSpell},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CounterObject{
							Object: game.TargetStackObjectReference(0),
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing: game.DelayedAtBeginningOfNextUpkeep,
								Content: game.AbilityContent{
									Modes: []game.Mode{
										game.Mode{
											Sequence: []game.Instruction{
												game.Instruction{
													Primitive: game.Choose{
														Choice: game.ResolutionChoice{
															Kind:            game.ResolutionChoiceNumber,
															Prompt:          "Choose how many cards to draw.",
															PlayerReference: func() *game.PlayerReference { ref := game.CapturedTargetControllerReference(0); return &ref }(),
															MaxNumber:       2,
														},
														PublishChoice: game.ChoiceKey("delayed-draw-amount"),
													},
												},
												game.Instruction{
													Primitive: game.Draw{
														Amount: game.Dynamic(game.DynamicAmount{
															Kind:      game.DynamicAmountChosenNumber,
															ResultKey: game.ResultKey("delayed-draw-amount"),
														}),
														Player: game.CapturedTargetControllerReference(0),
													},
												},
											},
										},
									},
									MinModes: 1,
									MaxModes: 1,
								},
							},
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing: game.DelayedAtBeginningOfNextUpkeep,
								Content: game.AbilityContent{
									Modes: []game.Mode{
										game.Mode{
											Sequence: []game.Instruction{
												game.Instruction{
													Primitive: game.Draw{
														Amount: game.Fixed(1),
														Player: game.ControllerReference(),
													},
												},
											},
										},
									},
									MinModes: 1,
									MaxModes: 1,
								},
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Counter target spell. Its controller may draw up to two cards at the beginning of the next turn's upkeep.
			You draw a card at the beginning of the next turn's upkeep.
		`,
		},
	}
}
