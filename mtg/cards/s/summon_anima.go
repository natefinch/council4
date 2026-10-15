package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SummonAnima is the card definition for Summon: Anima.
//
// Type: Enchantment Creature — Saga Horror
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter. Sacrifice after IV.)
//	I, II, III — Pain — You draw a card and you lose 1 life.
//	IV — Oblivion — Each opponent sacrifices a creature of their choice and loses 3 life.
//	Menace
var SummonAnima = newSummonAnima

func newSummonAnima() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Summon: Anima",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Saga, types.Horror},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I, II, III — Pain — You draw a card and you lose 1 life.",
					Chapters: []int{1, 2, 3},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.LoseLife{
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
					Text:     "IV — Oblivion — Each opponent sacrifices a creature of their choice and loses 3 life.",
					Chapters: []int{4},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.SacrificePermanents{
											PlayerGroup: game.OpponentsReference(),
											Amount:      game.Fixed(1),
											Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}},
										},
									},
									game.Instruction{
										Primitive: game.LoseLife{
											Amount:      game.Fixed(3),
											PlayerGroup: game.OpponentsReference(),
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
			(As this Saga enters and after your draw step, add a lore counter. Sacrifice after IV.)
			I, II, III — Pain — You draw a card and you lose 1 life.
			IV — Oblivion — Each opponent sacrifices a creature of their choice and loses 3 life.
			Menace
		`,
		},
	}
}
