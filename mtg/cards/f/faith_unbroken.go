package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FaithUnbroken is the card definition for Faith Unbroken.
//
// Type: Enchantment — Aura
// Cost: {3}{W}
//
// Oracle text:
//
//	Enchant creature you control
//	When this Aura enters, exile target creature an opponent controls until this Aura leaves the battlefield.
//	Enchanted creature gets +2/+2.
var FaithUnbroken = newFaithUnbroken()

func newFaithUnbroken() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Faith Unbroken",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature you control",
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypesAny: []types.Card{types.Creature},
						Controller:       game.ControllerYou,
					}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature an opponent controls",
								Allow:      game.TargetAllowPermanent,
								Selection: opt.Val(game.Selection{
									RequiredTypesAny: []types.Card{types.Creature},
									Controller:       game.ControllerOpponent,
								}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Exile{
									Object:         game.TargetPermanentReference(0),
									ExileLinkedKey: game.LinkedKey("exile-until-leaves"),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:         game.EventZoneChanged,
							Source:        game.TriggerSourceSelf,
							MatchFromZone: true,
							FromZone:      zone.Battlefield,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source: game.LinkedBattlefieldSource(game.LinkedKey("exile-until-leaves")),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant creature you control
			When this Aura enters, exile target creature an opponent controls until this Aura leaves the battlefield.
			Enchanted creature gets +2/+2.
		`,
		},
	}
}
