package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RevengeOfRavens is the card definition for Revenge of Ravens.
//
// Type: Enchantment
// Cost: {3}{B}
//
// Oracle text:
//
//	Whenever a creature attacks you or a planeswalker you control, that creature's controller loses 1 life and you gain 1 life.
var RevengeOfRavens = newRevengeOfRavens()

func newRevengeOfRavens() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Revenge of Ravens",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                    game.EventAttackerDeclared,
							Player:                   game.TriggerPlayerYou,
							AttackRecipient:          game.AttackRecipientPlayer | game.AttackRecipientPlaneswalker,
							SubjectSelection:         game.Selection{RequiredTypes: []types.Card{types.Creature}},
							AttackRecipientSelection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}, Controller: game.ControllerYou},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.LoseLife{
									Amount: game.Fixed(1),
									Player: game.ObjectControllerReference(game.EventPermanentReference()),
								},
							},
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever a creature attacks you or a planeswalker you control, that creature's controller loses 1 life and you gain 1 life.
		`,
		},
	}
}
