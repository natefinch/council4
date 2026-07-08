package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EdricSpymasterOfTrest is the card definition for Edric, Spymaster of Trest.
//
// Type: Legendary Creature — Elf Rogue
// Cost: {1}{G}{U}
//
// Oracle text:
//
//	Whenever a creature deals combat damage to one of your opponents, its controller may draw a card.
var EdricSpymasterOfTrest = newEdricSpymasterOfTrest

func newEdricSpymasterOfTrest() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Edric, Spymaster of Trest",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.U,
			}),
			Colors:     []color.Color{color.Green, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elf, types.Rogue},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                 game.EventDamageDealt,
							Subject:               game.TriggerSubjectDamageSource,
							Player:                game.TriggerPlayerOpponent,
							RequireCombatDamage:   true,
							DamageRecipient:       game.DamageRecipientPlayer,
							DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ObjectControllerReference(game.EventPermanentReference()),
								},
								Optional:      true,
								OptionalActor: opt.Val(game.ObjectControllerReference(game.EventPermanentReference())),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever a creature deals combat damage to one of your opponents, its controller may draw a card.
		`,
		},
	}
}
