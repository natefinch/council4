package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DemonicAppetite is the card definition for Demonic Appetite.
//
// Type: Enchantment — Aura
// Cost: {B}
//
// Oracle text:
//
//	Enchant creature you control
//	Enchanted creature gets +3/+3.
//	At the beginning of your upkeep, sacrifice a creature.
var DemonicAppetite = newDemonicAppetite()

func newDemonicAppetite() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Demonic Appetite",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature you control",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{types.Creature},
						Controller:     game.ControllerYou,
					},
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     3,
							ToughnessDelta: 3,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Amount:    game.Fixed(1),
									Player:    game.ControllerReference(),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant creature you control
			Enchanted creature gets +3/+3.
			At the beginning of your upkeep, sacrifice a creature.
		`,
		},
	}
}
