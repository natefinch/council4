package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MesmericOrb is the card definition for Mesmeric Orb.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	Whenever a permanent becomes untapped, that permanent's controller mills a card.
var MesmericOrb = newMesmericOrb()

func newMesmericOrb() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Mesmeric Orb",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event: game.EventPermanentUntapped,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Mill{
									Amount: game.Fixed(1),
									Player: game.ObjectControllerReference(game.EventPermanentReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever a permanent becomes untapped, that permanent's controller mills a card.
		`,
		},
	}
}
