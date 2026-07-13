package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CrystallineNautilus is the card definition for Crystalline Nautilus.
//
// Type: Enchantment Creature — Nautilus
// Cost: {2}{U}
//
// Oracle text:
//
//	Bestow {3}{U}{U} (If you cast this card for its bestow cost, it's an Aura spell with enchant creature. It becomes a creature again if it's not attached.)
//	When this permanent becomes the target of a spell or ability, sacrifice it.
//	Enchanted creature gets +4/+4 and has "When this creature becomes the target of a spell or ability, sacrifice it."
var CrystallineNautilus = newCrystallineNautilus

func newCrystallineNautilus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Crystalline Nautilus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Nautilus},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.BestowStaticAbility(cost.Mana{cost.O(3), cost.U, cost.U}, &game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     4,
							ToughnessDelta: 4,
						},
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.TriggeredAbility{
									Trigger: game.TriggerCondition{
										Type: game.TriggerWhen,
										Pattern: game.TriggerPattern{
											Event:  game.EventObjectBecameTarget,
											Source: game.TriggerSourceSelf,
										},
									},
									Content: game.Mode{
										Sequence: []game.Instruction{
											{
												Primitive: game.Sacrifice{
													Object: game.EventPermanentReference(),
												},
											},
										},
									}.Ability(),
								}),
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventObjectBecameTarget,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Sacrifice{
									Object: game.EventPermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Bestow {3}{U}{U} (If you cast this card for its bestow cost, it's an Aura spell with enchant creature. It becomes a creature again if it's not attached.)
			When this permanent becomes the target of a spell or ability, sacrifice it.
			Enchanted creature gets +4/+4 and has "When this creature becomes the target of a spell or ability, sacrifice it."
		`,
		},
	}
}
