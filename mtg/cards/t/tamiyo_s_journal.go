package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TamiyoSJournal is the card definition for Tamiyo's Journal.
//
// Type: Legendary Artifact — Book
// Cost: {5}
//
// Oracle text:
//
//	At the beginning of your upkeep, investigate. (Create a Clue token. It's an artifact with "{2}, Sacrifice this token: Draw a card.")
//	{T}, Sacrifice three Clues: Search your library for a card, put that card into your hand, then shuffle.
var TamiyoSJournal = newTamiyoSJournal

func newTamiyoSJournal() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Tamiyo's Journal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Book},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Sacrifice three Clues: Search your library for a card, put that card into your hand, then shuffle.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalSacrifice,
							Text:        "Sacrifice three Clues",
							Amount:      3,
							SubtypesAny: cost.SubtypeSet{0: types.Clue},
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Search{
											Player: game.ControllerReference(),
											Spec: game.SearchSpec{
												SourceZone:       zone.Library,
												Destination:      zone.Hand,
												FailToFindPolicy: game.SearchMustFindIfAvailable,
											},
											Amount: game.Fixed(1),
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your upkeep, investigate. (Create a Clue token. It's an artifact with \"{2}, Sacrifice this token: Draw a card.\")",
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
										Primitive: game.Investigate{
											Amount: game.Fixed(1),
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
			At the beginning of your upkeep, investigate. (Create a Clue token. It's an artifact with "{2}, Sacrifice this token: Draw a card.")
			{T}, Sacrifice three Clues: Search your library for a card, put that card into your hand, then shuffle.
		`,
		},
	}
}
