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

	card.StaticAbilities = append(card.StaticAbilities, game.StaticAbility{
		Text: `
				Equipped creature gets +2/+0 and has reach.
			`,
		ContinuousEffects: []game.ContinuousEffect{
			{
				Layer:      game.LayerPowerToughnessModify,
				Group:      game.AttachedObjectGroup(game.SourcePermanentReference()),
				PowerDelta: 2,
			},
			{
				Layer: game.LayerAbility,
				Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
				AddKeywords: []game.Keyword{
					game.Reach,
				},
			},
		},
	},
	)

	card.TriggeredAbilities = append(card.TriggeredAbilities,
		game.TriggeredAbility{
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
			Content: game.Mode{
				Targets: []game.TargetSpec{
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "player or planeswalker",
						Allow:      game.TargetAllowPlayer | game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{
							RequiredTypesAny: []types.Card{
								types.Planeswalker,
							},
						}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:   game.DynamicAmountObjectPower,
								Object: game.EventPermanentReference(),
							}),
							Recipient:    game.AnyTargetDamageRecipient(0),
							DamageSource: opt.Val(game.EventPermanentReference()),
						},
					},
				},
			}.Ability(),
		},
	)

	card.ActivatedAbilities = append(card.ActivatedAbilities,
		game.EquipActivatedAbility(cost.Mana{cost.O(2), cost.R}),
	)
	return card
}
