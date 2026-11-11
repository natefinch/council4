package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var MizziumSkin = cardset.Entry{
	OracleID: "02d998a6-0afd-4be1-a209-05618fcc43ea",
	Name:     "Mizzium Skin",
	New:      newMizziumSkin,
	Verification: "mtg/rules/authored_existing_paragraph_effects_test.go: " +
		"TestMizziumSkinScenario; CR 2026-11-06 601.2b/c/f, 608.2b, 611.2a/c, 612.1, 613.1f, 613.4c, 702.11a/b, 702.96a/b",
}

var BountyOfMight = cardset.Entry{
	OracleID: "6f067bfd-1c59-463f-99e1-61161b43b405",
	Name:     "Bounty of Might",
	New:      newBountyOfMight,
	Verification: "mtg/rules/authored_existing_paragraph_effects_test.go: " +
		"TestBountyOfMightScenario; CR 2026-11-06 601.2c, 608.2b, 611.2a, 613.4c",
}

var LongGoodbye = cardset.Entry{
	OracleID: "9d63ba8a-a5cd-4fdb-bba3-95612be0077c",
	Name:     "Long Goodbye",
	New:      newLongGoodbye,
	Verification: "mtg/rules/authored_existing_paragraph_effects_test.go: " +
		"TestLongGoodbyeScenario; CR 2026-11-06 101.2, 202.3, 601.2c, 608.2b, 701.6a, 701.8a, 702.21a",
}

func newMizziumSkin() *game.CardDef {
	const text = "Target creature you control gets +0/+1 and gains hexproof until end of turn.\nOverload {1}{U} (You may cast this spell for its overload cost. If you do, change \"target\" in its text to \"each.\")"
	creaturesYouControl := game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		Controller:    game.ControllerYou,
	}
	effects := []game.ContinuousEffect{
		{Layer: game.LayerPowerToughnessModify, ToughnessDelta: 1},
		{Layer: game.LayerAbility, AddKeywords: []game.Keyword{game.Hexproof}},
	}
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Mizzium Skin",
			ManaCost:   opt.Val(cost.Mana{cost.U}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Blue},
			OracleText: text,
			SpellAbility: opt.Val(game.Mode{
				Text: text,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature you control",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(creaturesYouControl),
				}},
				Sequence: []game.Instruction{{
					Primitive: game.ApplyContinuous{
						Object:            opt.Val(game.TargetPermanentReference(0)),
						ContinuousEffects: effects,
						Duration:          game.DurationUntilEndOfTurn,
					},
				}},
			}.Ability()),
			Overload: opt.Val(game.OverloadAbility{
				Cost: cost.Mana{cost.O(1), cost.U},
				SpellAbility: game.Mode{
					Sequence: []game.Instruction{{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								{
									Group:          game.BattlefieldGroup(creaturesYouControl),
									Layer:          game.LayerPowerToughnessModify,
									ToughnessDelta: 1,
								},
								{
									Group:       game.BattlefieldGroup(creaturesYouControl),
									Layer:       game.LayerAbility,
									AddKeywords: []game.Keyword{game.Hexproof},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					}},
				}.Ability(),
			}),
		},
	}
}

func newBountyOfMight() *game.CardDef {
	const text = "Target creature gets +3/+3 until end of turn.\nTarget creature gets +3/+3 until end of turn.\nTarget creature gets +3/+3 until end of turn."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Bounty of Might",
			ManaCost:   opt.Val(cost.Mana{cost.O(4), cost.G, cost.G}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Green},
			OracleText: text,
			SpellAbility: opt.Val(game.Mode{
				Text:    text,
				Targets: []game.TargetSpec{creatureTarget(), creatureTarget(), creatureTarget()},
				Sequence: []game.Instruction{
					{Primitive: game.ModifyPT{
						Object: game.TargetPermanentReference(0), PowerDelta: game.Fixed(3),
						ToughnessDelta: game.Fixed(3), Duration: game.DurationUntilEndOfTurn,
					}},
					{Primitive: game.ModifyPT{
						Object: game.TargetPermanentReference(1), PowerDelta: game.Fixed(3),
						ToughnessDelta: game.Fixed(3), Duration: game.DurationUntilEndOfTurn,
					}},
					{Primitive: game.ModifyPT{
						Object: game.TargetPermanentReference(2), PowerDelta: game.Fixed(3),
						ToughnessDelta: game.Fixed(3), Duration: game.DurationUntilEndOfTurn,
					}},
				},
			}.Ability()),
		},
	}
}

func newLongGoodbye() *game.CardDef {
	const text = "This spell can't be countered. (This includes by the ward ability.)\nDestroy target creature or planeswalker with mana value 3 or less."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:       "Long Goodbye",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.B}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Black},
			OracleText: text,
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Text: text,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature or planeswalker with mana value 3 or less",
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker},
						ManaValue: opt.Val(compare.Int{
							Op:    compare.LessOrEqual,
							Value: 3,
						}),
					}),
				}},
				Sequence: []game.Instruction{{
					Primitive: game.Destroy{
						Object: game.TargetPermanentReference(0),
					},
				}},
			}.Ability()),
		},
	}
}
