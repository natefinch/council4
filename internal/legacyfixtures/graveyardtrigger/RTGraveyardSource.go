// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTGraveyardSource is the card definition for RT Graveyard Source.
//
// Type: Creature
//
// Oracle text:
//
//	At the beginning of your upkeep, you may return this card from your graveyard to your hand. You gain 1 life.
var RTGraveyardSource = newRTGraveyardSource

func newRTGraveyardSource() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "RT Graveyard Source",
			Types:     []types.Card{types.Creature},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					ZoneOfFunction: zone.Graveyard,
					Text:           "At the beginning of your upkeep, you may return this card from your graveyard to your hand. You gain 1 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepUpkeep,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceSource,
											},
											FromZone:    zone.Graveyard,
											Destination: zone.Hand,
										},
										Optional: true,
									},
									game.Instruction{
										Primitive: game.GainLife{
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
			At the beginning of your upkeep, you may return this card from your graveyard to your hand. You gain 1 life.
		`,
		},
	}
}
