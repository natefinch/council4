package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CaptureSphere is the card definition for Capture Sphere.
//
// Type: Enchantment — Aura
// Cost: {3}{U}
//
// Oracle text:
//
//	Flash (You may cast this spell any time you could cast an instant.)
//	Enchant creature
//	When this Aura enters, tap enchanted creature.
//	Enchanted creature doesn't untap during its controller's untap step.
var CaptureSphere = newCaptureSphere()

func newCaptureSphere() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Capture Sphere",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
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
			Flash (You may cast this spell any time you could cast an instant.)
			Enchant creature
			When this Aura enters, tap enchanted creature.
			Enchanted creature doesn't untap during its controller's untap step.
		`,
		},
	}
}
