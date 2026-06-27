package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// KamahlHeartOfKrosa is the card definition for Kamahl, Heart of Krosa.
//
// Type: Legendary Creature — Human Druid
// Cost: {6}{G}{G}
//
// Oracle text:
//
//	At the beginning of combat on your turn, creatures you control get +3/+3 and gain trample until end of turn.
//	{1}{G}: Until end of turn, target land you control becomes a 1/1 Elemental creature with vigilance, indestructible, and haste. It's still a land.
//	Partner (You can have two commanders if both have partner.)
var KamahlHeartOfKrosa = newKamahlHeartOfKrosa()

func newKamahlHeartOfKrosa() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Kamahl, Heart of Krosa",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Druid},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.PartnerStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{1}{G}: Until end of turn, target land you control becomes a 1/1 Elemental creature with vigilance, indestructible, and haste. It's still a land.",
					ManaCost:       opt.Val(cost.Mana{cost.O(1), cost.G}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target land you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:       game.LayerType,
											AddTypes:    []types.Card{types.Creature},
											AddSubtypes: []types.Sub{types.Elemental},
										},
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Vigilance,
												game.Indestructible,
												game.Haste,
											},
										},
										game.ContinuousEffect{
											Layer:        game.LayerPowerToughnessSet,
											SetPower:     opt.Val(game.PT{Value: 1}),
											SetToughness: opt.Val(game.PT{Value: 1}),
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepBeginningOfCombat,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:          game.LayerPowerToughnessModify,
											Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											PowerDelta:     3,
											ToughnessDelta: 3,
										},
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											AddKeywords: []game.Keyword{
												game.Trample,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of combat on your turn, creatures you control get +3/+3 and gain trample until end of turn.
			{1}{G}: Until end of turn, target land you control becomes a 1/1 Elemental creature with vigilance, indestructible, and haste. It's still a land.
			Partner (You can have two commanders if both have partner.)
		`,
		},
	}
}
