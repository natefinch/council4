package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const thoughtlaceOracleText = "Target spell or permanent becomes blue. (Mana symbols on that permanent remain unchanged.)"

var Thoughtlace = cardset.Entry{
	OracleID: "6452b6a6-6235-46a3-a712-a26592450438",
	Name:     "Thoughtlace",
	New:      newThoughtlace,
	Verification: "mtg/rules/color_laces_test.go: TestThoughtlaceTargetPermanentBecomesBlue; " +
		"CR 2026-11-06 105.3, 115.1a, 202.2, 400.7a, 601.2c/f/h, 608.2b, 611.2a/c, 613.1e",
}

func newThoughtlace() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Thoughtlace",
			OracleText: thoughtlaceOracleText,
			ManaCost:   opt.Val(cost.Mana{cost.U}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Blue},
			SpellAbility: opt.Val(game.Mode{
				Text: thoughtlaceOracleText,
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
							SetColors: []color.Color{color.Blue},
						}},
						Duration: game.DurationPermanent,
					},
				}},
			}.Ability()),
		},
	}
}
