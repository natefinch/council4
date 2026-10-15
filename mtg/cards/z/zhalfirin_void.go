package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// ZhalfirinVoid is the card definition for Zhalfirin Void.
//
// Type: Land
//
// Oracle text:
//
//	When this land enters, scry 1. (Look at the top card of your library. You may put that card on the bottom.)
//	{T}: Add {C}.
var ZhalfirinVoid = newZhalfirinVoid

func newZhalfirinVoid() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Zhalfirin Void",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this land enters, scry 1. (Look at the top card of your library. You may put that card on the bottom.)",
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
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Scry{
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
			When this land enters, scry 1. (Look at the top card of your library. You may put that card on the bottom.)
			{T}: Add {C}.
		`,
		},
	}
}
