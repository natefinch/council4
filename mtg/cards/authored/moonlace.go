package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const moonlaceOracleText = "Target spell or permanent becomes colorless."

var Moonlace = cardset.Entry{
	OracleID: "52d4dc8a-2799-4b4e-9c20-f14bb75c2279",
	Name:     "Moonlace",
	New:      newMoonlace,
	Verification: "mtg/rules/color_laces_test.go: TestMoonlaceTargetPermanentBecomesColorless; " +
		"CR 2026-11-06 105.3, 115.1a, 202.2, 400.7a, 601.2c/f/h, 608.2b, 611.2a/c, 613.1e",
}

func newMoonlace() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Moonlace",
			OracleText: moonlaceOracleText,
			ManaCost:   opt.Val(cost.Mana{cost.U}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Blue},
			SpellAbility: opt.Val(game.Mode{
				Text: moonlaceOracleText,
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
							Layer:        game.LayerColor,
							SetColorless: true,
						}},
						Duration: game.DurationPermanent,
					},
				}},
			}.Ability()),
		},
	}
}
