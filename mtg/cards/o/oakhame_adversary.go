package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OakhameAdversary is the card definition for Oakhame Adversary.
//
// Type: Creature — Elf Warrior
// Cost: {3}{G}
//
// Oracle text:
//
//	This spell costs {2} less to cast if an opponent controls a green permanent.
//	Deathtouch
//	Whenever this creature deals combat damage to a player, draw a card.
var OakhameAdversary = newOakhameAdversary()

func newOakhameAdversary() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Oakhame Adversary",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 2,
								ReductionCondition: opt.Val(game.Condition{
									AnyOpponentControls: opt.Val(game.SelectionCount{
										Selection: game.Selection{ColorsAny: []color.Color{color.Green}},
									}),
								}),
							},
						},
					},
				},
				game.DeathtouchStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:               game.EventDamageDealt,
							Source:              game.TriggerSourceSelf,
							Subject:             game.TriggerSubjectDamageSource,
							RequireCombatDamage: true,
							DamageRecipient:     game.DamageRecipientPlayer,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			This spell costs {2} less to cast if an opponent controls a green permanent.
			Deathtouch
			Whenever this creature deals combat damage to a player, draw a card.
		`,
		},
	}
}
