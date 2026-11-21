package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var WildRide = cardset.Entry{
	OracleID: "61978202-538b-4f92-acfc-e95f4a6ce466",
	Name:     "Wild Ride",
	New:      newWildRide,
	Verification: "mtg/rules/wild_ride_test.go: " +
		"TestWildRideHarmonizeTargetsCreatureAndGrantsHaste; printing " +
		"abc8c6f5-6135-428e-8476-1751f82623f9; Scryfall rulings retrieved " +
		"2026-11-21: 7; CR 2026-11-06 115.1a, 601.2a-i, 601.3, 307.1, " +
		"608.2b, 611.2a/c, 613.1f-g, 702.10b, 702.180a-b",
}

func newWildRide() *game.CardDef {
	const spellText = "Target creature gets +3/+0 and gains haste until end of turn."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Wild Ride",
			OracleText: spellText + "\nHarmonize {4}{R} (You may cast this card from your graveyard for its harmonize cost. You may tap a creature you control to reduce that cost by {X}, where X is its power. Then exile this spell.)",
			ManaCost:   opt.Val(cost.Mana{cost.R}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.Red},
			AlternativeCosts: []cost.Alternative{
				cost.HarmonizeAlternative(cost.Mana{cost.O(4), cost.R}),
			},
			SpellAbility: opt.Val(game.Mode{
				Text: spellText,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}},
				Sequence: []game.Instruction{
					{Primitive: game.ModifyPT{
						Object:     game.TargetPermanentReference(0),
						PowerDelta: game.Fixed(3),
						Duration:   game.DurationUntilEndOfTurn,
					}},
					{Primitive: game.ApplyContinuous{
						Object: opt.Val(game.TargetPermanentReference(0)),
						ContinuousEffects: []game.ContinuousEffect{{
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Haste},
						}},
						Duration: game.DurationUntilEndOfTurn,
					}},
				},
			}.Ability()),
		},
	}
}
