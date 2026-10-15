package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FuryStorm is the card definition for Fury Storm.
//
// Type: Instant
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	When you cast this spell, copy it for each time you've cast your commander from the command zone this game. You may choose new targets for the copies.
//	Copy target instant or sorcery spell. You may choose new targets for the copy.
var FuryStorm = newFuryStorm

func newFuryStorm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Fury Storm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
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
						Constraint: "target instant or sorcery spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},
							StackObjectKinds:  []game.StackObjectKind{game.StackSpell},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CopyStackObject{
							Object:              game.TargetStackObjectReference(0),
							Count:               1,
							MayChooseNewTargets: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			When you cast this spell, copy it for each time you've cast your commander from the command zone this game. You may choose new targets for the copies.
			Copy target instant or sorcery spell. You may choose new targets for the copy.
		`,
		},
	}
}
