package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IchorWellspring is the card definition for Ichor Wellspring.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	When this artifact enters or is put into a graveyard from the battlefield, draw a card.
var IchorWellspring = newIchorWellspring

func newIchorWellspring() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Ichor Wellspring",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:      game.EventPermanentEnteredBattlefield,
							Source:     game.TriggerSourceSelf,
							UnionEvent: game.EventPermanentDied,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this artifact enters or is put into a graveyard from the battlefield, draw a card.
		`,
		},
	}
}
