package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const lifelaceOracleText = "Target spell or permanent becomes green. (Mana symbols on that permanent remain unchanged.)"

var Lifelace = cardset.Entry{
	OracleID: "eec1de80-4b3d-481d-a235-c299e0381830",
	Name:     "Lifelace",
	New:      newLifelace,
	Verification: "mtg/rules/color_laces_test.go: TestLifelaceTargetPermanentBecomesGreen; " +
		"CR 2026-11-06 105.3, 115.1a, 202.2, 400.7a, 601.2c/f/h, 608.2b, 611.2a/c, 613.1e",
}

func newLifelace() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Lifelace",
			OracleText: lifelaceOracleText,
			ManaCost:   opt.Val(cost.Mana{cost.G}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Green},
			SpellAbility: opt.Val(game.Mode{
				Text: lifelaceOracleText,
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
							SetColors: []color.Color{color.Green},
						}},
						Duration: game.DurationPermanent,
					},
				}},
			}.Ability()),
		},
	}
}
