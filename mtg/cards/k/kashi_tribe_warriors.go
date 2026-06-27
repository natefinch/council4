package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KashiTribeWarriors is the card definition for Kashi-Tribe Warriors.
//
// Type: Creature — Snake Warrior
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	Whenever this creature deals combat damage to a creature, tap that creature and it doesn't untap during its controller's next untap step.
var KashiTribeWarriors = newKashiTribeWarriors()

func newKashiTribeWarriors() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Kashi-Tribe Warriors",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Snake, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                game.EventDamageDealt,
							Source:               game.TriggerSourceSelf,
							Subject:              game.TriggerSubjectDamageSource,
							RequireCombatDamage:  true,
							DamageRecipient:      game.DamageRecipientPermanent,
							DamageRecipientTypes: []types.Card{types.Creature},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Tap{
									Object: game.EventPermanentReference(),
								},
							},
							{
								Primitive: game.SkipNextUntap{
									Object: game.EventPermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature deals combat damage to a creature, tap that creature and it doesn't untap during its controller's next untap step.
		`,
		},
	}
}
