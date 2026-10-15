package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EchoStorm is the card definition for Echo Storm.
//
// Type: Sorcery
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	When you cast this spell, copy it for each time you've cast your commander from the command zone this game. You may choose new targets for the copies.
//	Create a token that's a copy of target artifact.
var EchoStorm = newEchoStorm

func newEchoStorm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Echo Storm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When you cast this spell, copy it for each time you've cast your commander from the command zone this game. You may choose new targets for the copies.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							Source:          game.TriggerSourceSelf,
							SelfWasCast:     true,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CopyStackObject{
											Object: game.EventStackObjectReference(),
											DynamicCount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountCommanderCastCount,
												Multiplier: 1,
											}),
											MayChooseNewTargets: true,
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
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target artifact",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenCopyOf(game.TokenCopySpec{
								Source: game.TokenCopySourceObject,
								Object: game.TargetPermanentReference(0),
							}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			When you cast this spell, copy it for each time you've cast your commander from the command zone this game. You may choose new targets for the copies.
			Create a token that's a copy of target artifact.
		`,
		},
	}
}
