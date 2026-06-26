package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FracturedSanity is the card definition for Fractured Sanity.
//
// Type: Sorcery
// Cost: {U}{U}{U}
//
// Oracle text:
//
//	Each opponent mills fourteen cards.
//	Cycling {1}{U} ({1}{U}, Discard this card: Draw a card.)
//	When you cycle this card, each opponent mills four cards.
var FracturedSanity = newFracturedSanity()

func newFracturedSanity() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Fractured Sanity",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(1), cost.U}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventCycled,
							Source: game.TriggerSourceSelf,
							Player: game.TriggerPlayerYou,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Mill{
									Amount:      game.Fixed(4),
									PlayerGroup: game.OpponentsReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Mill{
							Amount:      game.Fixed(14),
							PlayerGroup: game.OpponentsReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each opponent mills fourteen cards.
			Cycling {1}{U} ({1}{U}, Discard this card: Draw a card.)
			When you cycle this card, each opponent mills four cards.
		`,
		},
	}
}
