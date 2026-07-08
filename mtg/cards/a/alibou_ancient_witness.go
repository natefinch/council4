package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AlibouAncientWitness is the card definition for Alibou, Ancient Witness.
//
// Type: Legendary Artifact Creature — Golem
// Cost: {3}{R}{W}
//
// Oracle text:
//
//	Other artifact creatures you control have haste.
//	Whenever one or more artifact creatures you control attack, Alibou deals X damage to any target and you scry X, where X is the number of tapped artifacts you control.
var AlibouAncientWitness = newAlibouAncientWitness

func newAlibouAncientWitness() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Alibou, Ancient Witness",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.W,
			}),
			Colors:     []color.Color{color.Red, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Golem},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}}, game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Haste,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventAttackerDeclared,
							Controller:       game.TriggerControllerYou,
							OneOrMore:        true,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Artifact, types.Creature}},
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "any target",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind: game.DynamicAmountX,
									}),
									Recipient: game.AnyTargetDamageRecipient(0),
								},
							},
							{
								Primitive: game.Scry{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountSelector,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerYou, Tapped: game.TriTrue}),
									}),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Other artifact creatures you control have haste.
			Whenever one or more artifact creatures you control attack, Alibou deals X damage to any target and you scry X, where X is the number of tapped artifacts you control.
		`,
		},
	}
}
