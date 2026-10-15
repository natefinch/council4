package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GethSGrimoire is the card definition for Geth's Grimoire.
//
// Type: Artifact — Book
// Cost: {4}
//
// Oracle text:
//
//	Whenever an opponent discards a card, you may draw a card.
var GethSGrimoire = newGethSGrimoire

func newGethSGrimoire() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Geth's Grimoire",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Book},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever an opponent discards a card, you may draw a card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventCardDiscarded,
							Player:          game.TriggerPlayerOpponent,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Optional: true,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Whenever an opponent discards a card, you may draw a card.
		`,
		},
	}
}
