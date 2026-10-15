package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AltarOfTheBrood is the card definition for Altar of the Brood.
//
// Type: Artifact
// Cost: {1}
//
// Oracle text:
//
//	Whenever another permanent you control enters, each opponent mills a card.
var AltarOfTheBrood = newAltarOfTheBrood

func newAltarOfTheBrood() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Altar of the Brood",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types: []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever another permanent you control enters, each opponent mills a card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Controller:      game.TriggerControllerYou,
							ExcludeSelf:     true,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveTopOfLibrary{
											Amount:      game.Fixed(1),
											PlayerGroup: game.OpponentsReference(),
											Destination: zone.Graveyard,
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
			Whenever another permanent you control enters, each opponent mills a card.
		`,
		},
	}
}
