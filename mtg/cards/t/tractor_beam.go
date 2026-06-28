package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TractorBeam is the card definition for Tractor Beam.
//
// Type: Enchantment — Aura
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	Enchant creature or Spacecraft
//	When this Aura enters, tap enchanted permanent.
//	You control enchanted permanent.
//	Enchanted permanent doesn't untap during its controller's untap step.
var TractorBeam = newTractorBeam()

func newTractorBeam() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Tractor Beam",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{AnyOf: []game.Selection{game.Selection{RequiredTypesAny: []types.Card{types.Creature}}, game.Selection{SubtypesAny: []types.Sub{types.Sub("Spacecraft")}}}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:         game.LayerControl,
							NewController: opt.Val(game.Player1),
							Group:         game.AttachedObjectGroup(game.SourcePermanentReference()),
						},
					},
				},
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectDoesntUntap,
							AffectedAttached: true,
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
						Sequence: []game.Instruction{
							{
								Primitive: game.Tap{
									Object: game.SourceAttachedPermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant creature or Spacecraft
			When this Aura enters, tap enchanted permanent.
			You control enchanted permanent.
			Enchanted permanent doesn't untap during its controller's untap step.
		`,
		},
	}
}
