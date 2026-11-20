package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const (
	leapOfFlameOracleID      = "cf3b1070-2c08-49b9-aa78-fe676fab38aa"
	leapOfFlameReplicateText = "Replicate {U}{R} (When you cast this spell, copy it for each time you paid its replicate cost. You may choose new targets for the copies.)"
	leapOfFlameSpellText     = "Target creature gets +1/+0 and gains flying and first strike until end of turn."
)

var LeapOfFlame = cardset.Entry{
	OracleID: leapOfFlameOracleID,
	Name:     "Leap of Flame",
	New:      newLeapOfFlame,
	Verification: "mtg/rules/leap_of_flame_test.go: TestLeapOfFlameReplicateBuffsThreeCreatures; " +
		"printing c011b432-ec39-44c8-bf2a-1c3fec32a576; Scryfall rulings 2026-11-19: 0; " +
		"CR 2026-11-06 115.1a, 601.2a-i, 603.2-4, 608.2b, 611.2a, 613.1f-g, 702.56a, 707.10/.10c",
}

func newLeapOfFlame() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name:       "Leap of Flame",
			OracleText: leapOfFlameReplicateText + "\n" + leapOfFlameSpellText,
			ManaCost:   opt.Val(cost.Mana{cost.U, cost.R}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Blue, color.Red},
			StaticAbilities: []game.StaticAbility{{
				Text: leapOfFlameReplicateText,
				KeywordAbilities: []game.KeywordAbility{
					game.ReplicateKeyword{Cost: cost.Mana{cost.U, cost.R}},
				},
			}},
			SpellAbility: opt.Val(game.Mode{
				Text: leapOfFlameSpellText,
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
						PowerDelta: game.Fixed(1),
						Duration:   game.DurationUntilEndOfTurn,
					}},
					{Primitive: game.ApplyContinuous{
						Object: opt.Val(game.TargetPermanentReference(0)),
						ContinuousEffects: []game.ContinuousEffect{{
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Flying, game.FirstStrike},
						}},
						Duration: game.DurationUntilEndOfTurn,
					}},
				},
			}.Ability()),
		},
	}
}
