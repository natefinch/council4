package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// QuestingBeast is the card definition for Questing Beast.
//
// Type: Legendary Creature — Beast
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Vigilance, deathtouch, haste
//	Questing Beast can't be blocked by creatures with power 2 or less.
//	Combat damage that would be dealt by creatures you control can't be prevented.
//	Whenever Questing Beast deals combat damage to an opponent, it deals that much damage to target planeswalker that player controls.
var QuestingBeast = newQuestingBeast

func newQuestingBeast() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Questing Beast",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Beast},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.DeathtouchStaticBody,
				game.HasteStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlockedByCreaturesWith,
							AffectedSource: true,
							BlockerRestriction: game.BlockerRestriction{
								Kind:  game.BlockerRestrictionPowerLessOrEqual,
								Power: 2,
							},
						},
					},
				},
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:              game.RuleEffectCombatDamageCantBePrevented,
							AffectedSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:               game.EventDamageDealt,
							Source:              game.TriggerSourceSelf,
							Subject:             game.TriggerSubjectDamageSource,
							Player:              game.TriggerPlayerOpponent,
							RequireCombatDamage: true,
							DamageRecipient:     game.DamageRecipientPlayer,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target planeswalker that player controls",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Planeswalker}, ControlledByEventPlayer: true}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountEventDamage,
										Multiplier: 1,
									}),
									Recipient:    game.AnyTargetDamageRecipient(0),
									DamageSource: opt.Val(game.EventPermanentReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Vigilance, deathtouch, haste
			Questing Beast can't be blocked by creatures with power 2 or less.
			Combat damage that would be dealt by creatures you control can't be prevented.
			Whenever Questing Beast deals combat damage to an opponent, it deals that much damage to target planeswalker that player controls.
		`,
		},
	}
}
