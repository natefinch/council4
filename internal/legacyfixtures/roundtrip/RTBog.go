// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// RTBog is the card definition for RT Bog.
//
// Type: Land
//
// Oracle text:
//
//	When this land enters, exile target player's graveyard.
var RTBog = newRTBog

func newRTBog() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Bog",
			Types: []types.Card{types.Land},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this land enters, exile target player's graveyard.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target player's graveyard",
										Allow:      game.TargetAllowPlayer,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Player:      game.TargetPlayerReference(0),
											FromZone:    zone.Graveyard,
											Destination: zone.Exile,
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
			When this land enters, exile target player's graveyard.
		`,
		},
	}
}
