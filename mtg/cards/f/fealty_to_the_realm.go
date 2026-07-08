package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FealtyToTheRealm is the card definition for Fealty to the Realm.
//
// Type: Enchantment — Aura
// Cost: {4}{U}
//
// Oracle text:
//
//	Enchant creature
//	When this Aura enters, you become the monarch.
//	The monarch controls enchanted creature.
//	Enchanted creature attacks each combat if able and can't attack you.
var FealtyToTheRealm = newFealtyToTheRealm

func newFealtyToTheRealm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Fealty to the Realm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}),
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:                  game.LayerControl,
							NewControllerIsMonarch: true,
							Group:                  game.AttachedObjectGroup(game.SourcePermanentReference()),
						},
					},
				},
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectMustAttack,
							AffectedAttached: true,
						},
						game.RuleEffect{
							Kind:                      game.RuleEffectCantAttack,
							AffectedAttached:          true,
							DefendingPlayer:           game.PlayerYou,
							DefendingPlayerDirectOnly: true,
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
								Primitive: game.BecomeMonarch{
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant creature
			When this Aura enters, you become the monarch.
			The monarch controls enchanted creature.
			Enchanted creature attacks each combat if able and can't attack you.
		`,
		},
	}
}
