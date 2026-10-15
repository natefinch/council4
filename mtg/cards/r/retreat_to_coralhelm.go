package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RetreatToCoralhelm is the card definition for Retreat to Coralhelm.
//
// Type: Enchantment
// Cost: {2}{U}
//
// Oracle text:
//
//	Landfall — Whenever a land you control enters, choose one —
//	• You may tap or untap target creature.
//	• Scry 1. (Look at the top card of your library. You may put that card on the bottom.)
var RetreatToCoralhelm = newRetreatToCoralhelm

func newRetreatToCoralhelm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Retreat to Coralhelm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Landfall — Whenever a land you control enters, choose one —\n• You may tap or untap target creature.\n• Scry 1. (Look at the top card of your library. You may put that card on the bottom.)",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Land}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "You may tap or untap target creature.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.TapOrUntap{
											Object: game.TargetPermanentReference(0),
										},
										Optional: true,
									},
								},
							},
							game.Mode{
								Text: "Scry 1. (Look at the top card of your library. You may put that card on the bottom.)",
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Scry{
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
			OracleText: `
			Landfall — Whenever a land you control enters, choose one —
			• You may tap or untap target creature.
			• Scry 1. (Look at the top card of your library. You may put that card on the bottom.)
		`,
		},
	}
}
