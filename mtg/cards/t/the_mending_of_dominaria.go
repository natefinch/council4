package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TheMendingOfDominaria is the card definition for The Mending of Dominaria.
//
// Type: Enchantment — Saga
// Cost: {3}{G}{G}
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
//	I, II — Mill two cards, then you may return a creature card from your graveyard to your hand.
//	III — Return all land cards from your graveyard to the battlefield, then shuffle your graveyard into your library.
var TheMendingOfDominaria = newTheMendingOfDominaria

func newTheMendingOfDominaria() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "The Mending of Dominaria",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Saga},
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I, II — Mill two cards, then you may return a creature card from your graveyard to your hand.",
					Chapters: []int{1, 2},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveTopOfLibrary{
											Amount:      game.Fixed(2),
											Player:      game.ControllerReference(),
											Destination: zone.Graveyard,
										},
									},
									game.Instruction{
										Primitive: game.ChooseFromZone{
											Player:     game.ControllerReference(),
											SourceZone: zone.Graveyard,
											Filter:     game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
											Quantity:   game.Fixed(1),
											Destination: game.ChooseDestination{
												Zone: zone.Hand,
											},
											Prompt: "Choose a card to return to your hand",
										},
										Optional: true,
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ChapterAbility{
					Text:     "III — Return all land cards from your graveyard to the battlefield, then shuffle your graveyard into your library.",
					Chapters: []int{3},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MassReturnFromGraveyard{
											Player:      game.ControllerReference(),
											Selection:   game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou},
											Destination: zone.Battlefield,
										},
									},
									game.Instruction{
										Primitive: game.ShuffleGraveyardIntoLibrary{
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
			(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
			I, II — Mill two cards, then you may return a creature card from your graveyard to your hand.
			III — Return all land cards from your graveyard to the battlefield, then shuffle your graveyard into your library.
		`,
		},
	}
}
