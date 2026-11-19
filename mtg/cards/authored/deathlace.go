package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const deathlaceOracleText = "Target spell or permanent becomes black. (Mana symbols on that permanent remain unchanged.)"

var Deathlace = cardset.Entry{
	OracleID: "fb80aaba-352a-4b58-8db2-1e02d542819c",
	Name:     "Deathlace",
	New:      newDeathlace,
	Verification: "mtg/rules/color_laces_test.go: TestDeathlaceTargetPermanentBecomesBlack; " +
		"CR 2026-11-06 105.3, 115.1a, 202.2, 400.7a, 601.2c/f/h, 608.2b, 611.2a/c, 613.1e",
}

func newDeathlace() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:       "Deathlace",
			OracleText: deathlaceOracleText,
			ManaCost:   opt.Val(cost.Mana{cost.B}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Black},
			SpellAbility: opt.Val(game.Mode{
				Text: deathlaceOracleText,
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
							SetColors: []color.Color{color.Black},
						}},
						Duration: game.DurationPermanent,
					},
				}},
			}.Ability()),
		},
	}
}
