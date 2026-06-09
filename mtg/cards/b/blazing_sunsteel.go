package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BlazingSunsteel is the card definition for Blazing Sunsteel.
//
// Type: Artifact — Equipment
// Cost: {1}{R}
//
// Oracle text:
//
//	Equipped creature gets +1/+0 for each opponent you have.
//	Whenever equipped creature is dealt damage, it deals that much damage to any target.
//	Equip {4}
var BlazingSunsteel = func() *game.CardDef {
	card := &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Blazing Sunsteel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			OracleText: `
				Equipped creature gets +1/+0 for each opponent you have.
				Whenever equipped creature is dealt damage, it deals that much damage to any target.
				Equip {4}
			`,
		},
	}

	card.StaticAbilities = append(card.StaticAbilities, game.StaticAbility{
		Text: `
				Equipped creature gets +1/+0 for each opponent you have.
			`,
		ContinuousEffects: []game.ContinuousEffect{
			{
				Layer: game.LayerPowerToughnessModify,
				Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
				PowerDeltaDynamic: opt.Val(game.DynamicAmount{
					Kind: game.DynamicAmountOpponentCount,
				}),
			},
		},
	},
	)

	card.TriggeredAbilities = append(card.TriggeredAbilities,
		game.TriggeredAbility{
			Text: `
				Whenever equipped creature is dealt damage, it deals that much damage to any target.
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
						Constraint: "any target",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountEventDamage,
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
		game.EquipActivatedAbility(cost.Mana{cost.O(4)}),
	)
	return card
}()
