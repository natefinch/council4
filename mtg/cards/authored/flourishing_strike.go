package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const flourishingStrikeOracleID = "98c44298-7b6a-4f98-bcb7-1a63a4c29f87"

var FlourishingStrike = cardset.Entry{
	OracleID: flourishingStrikeOracleID,
	Name:     "Flourishing Strike",
	New:      newFlourishingStrike,
	Verification: "mtg/rules/flourishing_strike_test.go: " +
		"TestFlourishingStrikeEntwineDealsDamageAndPumps",
}

func newFlourishingStrike() *game.CardDef {
	const damageText = "Flourishing Strike deals 5 damage to target creature with flying."
	const pumpText = "Target creature gets +3/+3 until end of turn."
	const text = "Choose one —\n• " + damageText + "\n• " + pumpText +
		"\nEntwine {2}{G} (Choose both if you pay the entwine cost.)"

	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Flourishing Strike",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.G}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Green},
			OracleText: text,
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					{
						Text: damageText,
						Targets: []game.TargetSpec{{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "target creature with flying",
							Allow:      game.TargetAllowPermanent,
							Selection: opt.Val(game.Selection{
								RequiredTypes: []types.Card{types.Creature},
								Keyword:       game.Flying,
							}),
						}},
						Sequence: []game.Instruction{{
							Primitive: game.Damage{
								Amount:    game.Fixed(5),
								Recipient: game.ObjectDamageRecipient(game.TargetPermanentReference(0)),
							},
						}},
					},
					{
						Text: pumpText,
						Targets: []game.TargetSpec{{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "target creature",
							Allow:      game.TargetAllowPermanent,
							Selection: opt.Val(game.Selection{
								RequiredTypes: []types.Card{types.Creature},
							}),
						}},
						Sequence: []game.Instruction{{
							Primitive: game.ApplyContinuous{
								Object: opt.Val(game.TargetPermanentReference(0)),
								ContinuousEffects: []game.ContinuousEffect{{
									Layer:          game.LayerPowerToughnessModify,
									PowerDelta:     3,
									ToughnessDelta: 3,
								}},
								Duration: game.DurationUntilEndOfTurn,
							},
						}},
					},
				},
				MinModes:     1,
				MaxModes:     2,
				EscalateCost: opt.Val(cost.Mana{cost.O(2), cost.G}),
			}),
		},
	}
}
