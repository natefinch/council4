package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Riddlekeeper is the card definition for Riddlekeeper.
//
// Type: Creature — Homunculus
// Cost: {2}{U}
//
// Oracle text:
//
//	Whenever a creature attacks you or a planeswalker you control, that creature's controller mills two cards.
var Riddlekeeper = newRiddlekeeper

func newRiddlekeeper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Riddlekeeper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Homunculus},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 4}),
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
								Primitive: game.Mill{
									Amount: game.Fixed(2),
									Player: game.ObjectControllerReference(game.EventPermanentReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever a creature attacks you or a planeswalker you control, that creature's controller mills two cards.
		`,
		},
	}
}
