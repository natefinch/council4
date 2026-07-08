package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KamiOfTheHonoredDead is the card definition for Kami of the Honored Dead.
//
// Type: Creature — Spirit
// Cost: {5}{W}{W}
//
// Oracle text:
//
//	Flying
//	Whenever this creature is dealt damage, you gain that much life.
//	Soulshift 6 (When this creature dies, you may return target Spirit card with mana value 6 or less from your graveyard to your hand.)
var KamiOfTheHonoredDead = newKamiOfTheHonoredDead

func newKamiOfTheHonoredDead() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Kami of the Honored Dead",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventDamageDealt,
							Source:          game.TriggerSourceSelf,
							Subject:         game.TriggerSubjectPermanent,
							DamageRecipient: game.DamageRecipientPermanent,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountEventDamage,
										Multiplier: 1,
									}),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.SoulshiftTriggeredAbility(6),
			},
			OracleText: `
			Flying
			Whenever this creature is dealt damage, you gain that much life.
			Soulshift 6 (When this creature dies, you may return target Spirit card with mana value 6 or less from your graveyard to your hand.)
		`,
		},
	}
}
