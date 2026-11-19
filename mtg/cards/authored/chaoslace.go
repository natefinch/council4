package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const chaoslaceOracleText = "Target spell or permanent becomes red. (Its mana symbols remain unchanged.)"

var Chaoslace = cardset.Entry{
	OracleID: "08842aa3-f923-46e9-a106-f542331e9cc1",
	Name:     "Chaoslace",
	New:      newChaoslace,
	Verification: "mtg/rules/chaoslace_test.go: TestChaoslaceTargetPermanentBecomesRed; " +
		"TestChaoslaceTargetSpellBecomesRed; CR 2026-11-06 105.3, 202.2, " +
		"400.7a, 601.2c/f/h, 608.2b, 611.2a/c, 613.1e",
}

func newChaoslace() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Chaoslace",
			OracleText: chaoslaceOracleText,
			ManaCost:   opt.Val(cost.Mana{cost.R}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Red},
			SpellAbility: opt.Val(game.Mode{
				Text: chaoslaceOracleText,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "spell or permanent",
					Allow:      game.TargetAllowStackObject | game.TargetAllowPermanent,
					Predicate:  game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}},
				}},
				Sequence: []game.Instruction{{
					Primitive: game.ApplyContinuous{
						Object: opt.Val(game.TargetObjectReference(0)),
						ContinuousEffects: []game.ContinuousEffect{{
							Layer:     game.LayerColor,
							SetColors: []color.Color{color.Red},
						}},
						Duration: game.DurationPermanent,
					},
				}},
			}.Ability()),
		},
	}
}
