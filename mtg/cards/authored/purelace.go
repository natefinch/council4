package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const purelaceOracleText = "Target spell or permanent becomes white. (Mana symbols on that permanent remain unchanged.)"

var Purelace = cardset.Entry{
	OracleID: "3773001a-8868-49ec-a406-298cf72359c2",
	Name:     "Purelace",
	New:      newPurelace,
	Verification: "mtg/rules/color_laces_test.go: TestPurelaceTargetPermanentBecomesWhite; " +
		"CR 2026-11-06 105.3, 115.1a, 202.2, 400.7a, 601.2c/f/h, 608.2b, 611.2a/c, 613.1e",
}

func newPurelace() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:       "Purelace",
			OracleText: purelaceOracleText,
			ManaCost:   opt.Val(cost.Mana{cost.W}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.White},
			SpellAbility: opt.Val(game.Mode{
				Text: purelaceOracleText,
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
							SetColors: []color.Color{color.White},
						}},
						Duration: game.DurationPermanent,
					},
				}},
			}.Ability()),
		},
	}
}
