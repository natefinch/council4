package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// OkibaReckonerRaid is the card definition for Okiba Reckoner Raid // Nezumi Road Captain.
//
// Type: Enchantment — Saga // Enchantment Creature — Rat Rogue
// Face: Nezumi Road Captain — Enchantment Creature — Rat Rogue
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter.)
//	I, II — Each opponent loses 1 life and you gain 1 life.
//	III — Exile this Saga, then return it to the battlefield transformed under your control.
var OkibaReckonerRaid = newOkibaReckonerRaid

func newOkibaReckonerRaid() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Okiba Reckoner Raid",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Saga},
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I, II — Each opponent loses 1 life and you gain 1 life.",
					Chapters: []int{1, 2},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.LoseLife{
											Amount:      game.Fixed(1),
											PlayerGroup: game.OpponentsReference(),
										},
										PublishResult: game.ResultKey("life-change"),
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
				game.ChapterAbility{
					Text:     "III — Exile this Saga, then return it to the battlefield transformed under your control.",
					Chapters: []int{3},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MovePermanent{
											Object:        game.SourcePermanentReference(),
											Destination:   zone.Exile,
											PublishLinked: game.LinkedKey("self-blink"),
										},
									},
									game.Instruction{
										Primitive: game.PutOnBattlefield{
											Source:           game.LinkedBattlefieldSource(game.LinkedKey("self-blink")),
											Recipient:        opt.Val(game.ControllerReference()),
											EntryTransformed: true,
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
			(As this Saga enters and after your draw step, add a lore counter.)
			I, II — Each opponent loses 1 life and you gain 1 life.
			III — Exile this Saga, then return it to the battlefield transformed under your control.
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:      "Nezumi Road Captain",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Rat, types.Rogue},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
				game.StaticAbility{
					Text: "Vehicles you control have menace. (They can't be blocked except by two or more creatures.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Vehicle")}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Menace},
						},
					},
				},
			},
			OracleText: `
			Menace
			Vehicles you control have menace. (They can't be blocked except by two or more creatures.)
		`,
		}),
	}
}
