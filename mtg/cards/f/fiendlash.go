package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Fiendlash is the card definition for Fiendlash.
//
// Type: Artifact — Equipment
// Cost: {1}{R}
//
// Oracle text:
//
//	Equipped creature gets +2/+0 and has reach.
//	Whenever equipped creature is dealt damage, it deals damage equal to its power to target player or planeswalker.
//	Equip {2}{R}
var Fiendlash = func() *game.CardDef {
	card := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Fiendlash",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			OracleText: `
				Equipped creature gets +2/+0 and has reach.
				Whenever equipped creature is dealt damage, it deals damage equal to its power to target player or planeswalker.
				Equip {2}{R}
			`,
		},
	}

	card.StaticAbilities = append(card.StaticAbilities,
		game.StaticAbilityBody{
			Text: `
				Equipped creature gets +2/+0 and has reach.
			`,
			Effects: []game.Effect{
				{
					Type:        game.EffectApplyContinuous,
					TargetIndex: game.TargetIndexSourcePermanent,
					ContinuousEffects: []game.ContinuousEffect{
						{
							Layer:          game.LayerPowerToughnessModify,
							Selector:       game.EffectSelectorEquippedCreature,
							PowerDelta:     2,
							ToughnessDelta: 0,
						},
						{
							Layer:       game.LayerAbility,
							Selector:    game.EffectSelectorEquippedCreature,
							AddKeywords: []game.Keyword{game.Reach},
						},
					},
				},
			},
		},
	)

	card.TriggeredAbilities = append(card.TriggeredAbilities,
		game.TriggeredAbilityBody{
			Text: `
				Whenever equipped creature is dealt damage, it deals damage equal to its power to target player or planeswalker.
			`,
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:           game.EventDamageDealt,
					Source:          game.TriggerSourceAttachedPermanent,
					DamageRecipient: game.DamageRecipientPermanent,
				},
			},
			Content: game.PlainAbilityContent{
				Targets: []game.TargetSpec{
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "player or planeswalker",
						Allow:      game.TargetAllowPlayer | game.TargetAllowPermanent,
						Predicate: game.TargetPredicate{
							PermanentTypes: []types.Card{types.Planeswalker},
						},
					},
				},
				Sequence: []game.Effect{
					{
						Type:        game.EffectDamage,
						TargetIndex: 0,
						DamageSource: opt.Val(game.ObjectReference{
							Kind:        game.ObjectReferenceAttachedPermanent,
							TargetIndex: game.TargetIndexSourcePermanent,
						}),
						DynamicAmount: opt.Val(game.DynamicAmount{
							Kind: game.DynamicAmountObjectPower,
							Object: game.ObjectReference{
								Kind:        game.ObjectReferenceAttachedPermanent,
								TargetIndex: game.TargetIndexSourcePermanent,
							},
						}),
					},
				},
			},
		},
	)

	card.ActivatedAbilities = append(card.ActivatedAbilities,
		game.ActivatedAbilityBody{
			Text: `
				Equip {2}{R}
			`,
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Timing: game.SorceryOnly,
			Content: game.PlainAbilityContent{
				Targets: []game.TargetSpec{
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "creature you control",
						Allow:      game.TargetAllowPermanent,
						Predicate: game.TargetPredicate{
							PermanentTypes: []types.Card{types.Creature},
							Controller:     game.ControllerYou,
						},
					},
				},
			},
			KeywordAbilities: []game.KeywordAbility{
				game.EquipKeyword{Cost: cost.Mana{
					cost.O(2),
					cost.R,
				}},
			},
		},
	)
	return card
}()
