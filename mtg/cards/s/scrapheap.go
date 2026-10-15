package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Scrapheap is the card definition for Scrapheap.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	Whenever an artifact or enchantment is put into your graveyard from the battlefield, you gain 1 life.
var Scrapheap = newScrapheap

func newScrapheap() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Scrapheap",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever an artifact or enchantment is put into your graveyard from the battlefield, you gain 1 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventZoneChanged,
							Player:           game.TriggerPlayerYou,
							SubjectSelection: game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}},
							MatchFromZone:    true,
							FromZone:         zone.Battlefield,
							MatchToZone:      true,
							ToZone:           zone.Graveyard,
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
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
			Whenever an artifact or enchantment is put into your graveyard from the battlefield, you gain 1 life.
		`,
		},
	}
}
