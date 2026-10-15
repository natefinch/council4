package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TheCurseOfFenric is the card definition for The Curse of Fenric.
//
// Type: Enchantment — Saga
// Cost: {2}{G}{W}
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
//	I — For each player, destroy up to one target creature that player controls. For each creature destroyed this way, its controller creates a 3/3 green Mutant creature token with deathtouch.
//	II — Target nontoken creature becomes a 6/6 legendary Horror creature named Fenric and loses all abilities.
//	III — Target Mutant fights another target creature named Fenric.
var TheCurseOfFenric = newTheCurseOfFenric

func newTheCurseOfFenric() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "The Curse of Fenric",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.W,
			}),
			Colors:   []color.Color{color.Green, color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Saga},
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I — For each player, destroy up to one target creature that player controls. For each creature destroyed this way, its controller creates a 3/3 green Mutant creature token with deathtouch.",
					Chapters: []int{1},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ForEachPlayer{
											Scope:     game.AllPlayersReference(),
											Chooser:   game.ControllerReference(),
											Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
											Removal:   game.DistributiveRemovalDestroy,
											LinkedKey: game.LinkedKey("destroyed-for-each-player"),
										},
									},
									game.Instruction{
										Primitive: game.CreateTokenForEachDestroyed{
											Source:    game.TokenDef(theCurseOfFenricToken),
											LinkedKey: game.LinkedKey("destroyed-for-each-player"),
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
					Text:     "II — Target nontoken creature becomes a 6/6 legendary Horror creature named Fenric and loses all abilities.",
					Chapters: []int{2},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target nontoken creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, NonToken: true}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.TargetPermanentReference(0)),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:           game.DurationPermanent,
													Layer:              game.LayerAbility,
													RemoveAllAbilities: true,
												},
												game.ContinuousEffect{
													Duration:      game.DurationPermanent,
													Layer:         game.LayerType,
													AddSupertypes: []types.Super{types.Legendary},
													SetTypes:      []types.Card{types.Creature},
													SetSubtypes:   []types.Sub{types.Horror},
												},
												game.ContinuousEffect{
													Duration: game.DurationPermanent,
													Layer:    game.LayerText,
													SetName:  "Fenric",
												},
												game.ContinuousEffect{
													Duration: game.DurationPermanent,
													Layer:    game.LayerPowerToughnessSet,
													SetPower: opt.Val(game.PT{
														Value: 6,
													}),
													SetToughness: opt.Val(game.PT{
														Value: 6,
													}),
												},
											},
											Duration: game.DurationPermanent,
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
					Text:     "III — Target Mutant fights another target creature named Fenric.",
					Chapters: []int{3},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "Target Mutant",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Mutant}}),
									},
									game.TargetSpec{
										MinTargets:               1,
										MaxTargets:               1,
										Constraint:               "another target creature named Fenric",
										Allow:                    game.TargetAllowPermanent,
										Selection:                opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Name: "Fenric"}),
										DistinctFromPriorTargets: true,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Fight{
											Object:        game.TargetPermanentReference(0),
											RelatedObject: game.TargetPermanentReference(1),
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
			I — For each player, destroy up to one target creature that player controls. For each creature destroyed this way, its controller creates a 3/3 green Mutant creature token with deathtouch.
			II — Target nontoken creature becomes a 6/6 legendary Horror creature named Fenric and loses all abilities.
			III — Target Mutant fights another target creature named Fenric.
		`,
		},
	}
}

var theCurseOfFenricToken = newTheCurseOfFenricToken()

func newTheCurseOfFenricToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Mutant",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Mutant},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
			},
		},
	}
}
