package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Mindcrank is the card definition for Mindcrank.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	Whenever an opponent loses life, that player mills that many cards. (Damage causes loss of life.)
var Mindcrank = newMindcrank()

func newMindcrank() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Mindcrank",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventLifeLost,
							Player: game.TriggerPlayerOpponent,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Mill{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountEventLifeChange,
										Multiplier: 1,
									}),
									Player: game.EventPlayerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever an opponent loses life, that player mills that many cards. (Damage causes loss of life.)
		`,
		},
	}
}
