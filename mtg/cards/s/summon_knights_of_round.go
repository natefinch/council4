package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SummonKnightsOfRound is the card definition for Summon: Knights of Round.
//
// Type: Enchantment Creature — Saga Knight
// Cost: {6}{W}{W}
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter. Sacrifice after V.)
//	I, II, III, IV — Create three 2/2 white Knight creature tokens.
//	V — Ultimate End — Other creatures you control get +2/+2 until end of turn. Put an indestructible counter on each of them.
//	Indestructible
var SummonKnightsOfRound = newSummonKnightsOfRound

func newSummonKnightsOfRound() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Summon: Knights of Round",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Saga, types.Knight},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
			},
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I, II, III, IV — Create three 2/2 white Knight creature tokens.",
					Chapters: []int{1, 2, 3, 4},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(3),
											Source: game.TokenDef(summonKnightsOfRoundToken),
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
					Text:     "V — Ultimate End — Other creatures you control get +2/+2 until end of turn. Put an indestructible counter on each of them.",
					Chapters: []int{5},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:       game.DurationPermanent,
													Group:          game.BattlefieldGroupExcluding(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}, game.SourcePermanentReference()),
													Layer:          game.LayerPowerToughnessModify,
													PowerDelta:     2,
													ToughnessDelta: 2,
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Group:       game.BattlefieldGroupExcluding(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}, game.SourcePermanentReference()),
											CounterKind: counter.Indestructible,
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
			(As this Saga enters and after your draw step, add a lore counter. Sacrifice after V.)
			I, II, III, IV — Create three 2/2 white Knight creature tokens.
			V — Ultimate End — Other creatures you control get +2/+2 until end of turn. Put an indestructible counter on each of them.
			Indestructible
		`,
		},
	}
}

var summonKnightsOfRoundToken = newSummonKnightsOfRoundToken()

func newSummonKnightsOfRoundToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Knight",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
