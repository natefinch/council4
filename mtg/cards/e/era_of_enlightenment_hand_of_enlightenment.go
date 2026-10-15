package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EraOfEnlightenment is the card definition for Era of Enlightenment // Hand of Enlightenment.
//
// Type: Enchantment — Saga // Enchantment Creature — Human Monk
// Face: Hand of Enlightenment — Enchantment Creature — Human Monk
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter.)
//	I — Scry 2.
//	II — You gain 2 life.
//	III — Exile this Saga, then return it to the battlefield transformed under your control.
var EraOfEnlightenment = newEraOfEnlightenment

func newEraOfEnlightenment() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Era of Enlightenment",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Saga},
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I — Scry 2.",
					Chapters: []int{1},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Scry{
											Amount: game.Fixed(2),
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
					Text:     "II — You gain 2 life.",
					Chapters: []int{2},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(2),
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
			I — Scry 2.
			II — You gain 2 life.
			III — Exile this Saga, then return it to the battlefield transformed under your control.
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:      "Hand of Enlightenment",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Monk},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
			},
			OracleText: `
			First strike
		`,
		}),
	}
}
