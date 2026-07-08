package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FlintGolem is the card definition for Flint Golem.
//
// Type: Artifact Creature — Golem
// Cost: {4}
//
// Oracle text:
//
//	Whenever this creature becomes blocked, defending player mills three cards.
var FlintGolem = newFlintGolem

func newFlintGolem() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Flint Golem",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerBecameBlocked,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Mill{
									Amount: game.Fixed(3),
									Player: game.DefendingPlayerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature becomes blocked, defending player mills three cards.
		`,
		},
	}
}
