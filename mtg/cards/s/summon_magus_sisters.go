package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SummonMagusSisters is the card definition for Summon: Magus Sisters.
//
// Type: Enchantment Creature — Saga Faerie
// Cost: {4}{G}
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
//	I, II, III — Choose one at random —
//	• Combine Powers! — Put three +1/+1 counters on target creature.
//	• Defense! — Put a shield counter on target creature. You gain 3 life.
//	• Fight! — This creature fights up to one target creature an opponent controls.
//	Haste
var SummonMagusSisters = newSummonMagusSisters

func newSummonMagusSisters() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Summon: Magus Sisters",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Saga, types.Faerie},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I, II, III — Choose one at random —\n• Combine Powers! — Put three +1/+1 counters on target creature.\n• Defense! — Put a shield counter on target creature. You gain 3 life.\n• Fight! — This creature fights up to one target creature an opponent controls.",
					Chapters: []int{1, 2, 3},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Combine Powers! — Put three +1/+1 counters on target creature.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(3),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.PlusOnePlusOne,
										},
									},
								},
							},
							game.Mode{
								Text: "Defense! — Put a shield counter on target creature. You gain 3 life.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.Shield,
										},
									},
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(3),
											Player: game.ControllerReference(),
										},
									},
								},
							},
							game.Mode{
								Text: "Fight! — This creature fights up to one target creature an opponent controls.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MaxTargets: 1,
										Constraint: "up to one target creature an opponent controls",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Fight{
											Object:        game.SourcePermanentReference(),
											RelatedObject: game.TargetPermanentReference(0),
										},
									},
								},
							},
						},
						MinModes:    1,
						MaxModes:    1,
						RandomModes: true,
					},
				},
			},
			OracleText: `
			(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
			I, II, III — Choose one at random —
			• Combine Powers! — Put three +1/+1 counters on target creature.
			• Defense! — Put a shield counter on target creature. You gain 3 life.
			• Fight! — This creature fights up to one target creature an opponent controls.
			Haste
		`,
		},
	}
}
