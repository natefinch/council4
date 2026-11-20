package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const rainOfRustOracleID = "074e467e-f574-4b44-b5a3-798b592a5a87"

var (
	rainOfRustArtifactProgram = game.ProgramReference{ID: rainOfRustOracleID + "/spell/0", Version: "1"}
	rainOfRustLandProgram     = game.ProgramReference{ID: rainOfRustOracleID + "/spell/1", Version: "1"}
)

var RainOfRust = cardset.Entry{
	OracleID: rainOfRustOracleID,
	Name:     "Rain of Rust",
	New:      newRainOfRust,
	Programs: []cardprogram.Program{
		{Reference: rainOfRustArtifactProgram, Resolve: resolveRainOfRustArtifact},
		{Reference: rainOfRustLandProgram, Resolve: resolveRainOfRustLand},
	},
	Verification: "mtg/rules/rain_of_rust_test.go: TestRainOfRustEntwineDestroysArtifactAndLand",
}

func newRainOfRust() *game.CardDef {
	const text = "Choose one —\n• Destroy target artifact.\n• Destroy target land.\nEntwine {3}{R} (Choose both if you pay the entwine cost.)"
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Rain of Rust",
			OracleText: text,
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.R, cost.R}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Red},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					{
						Text:    "Destroy target artifact.",
						Program: rainOfRustArtifactProgram,
						Targets: []game.TargetSpec{{
							MinTargets: 1, MaxTargets: 1, Constraint: "target artifact",
							Allow: game.TargetAllowPermanent,
							Selection: opt.Val(game.Selection{
								RequiredTypes: []types.Card{types.Artifact},
							}),
						}},
					},
					{
						Text:    "Destroy target land.",
						Program: rainOfRustLandProgram,
						Targets: []game.TargetSpec{{
							MinTargets: 1, MaxTargets: 1, Constraint: "target land",
							Allow: game.TargetAllowPermanent,
							Selection: opt.Val(game.Selection{
								RequiredTypes: []types.Card{types.Land},
							}),
						}},
					},
				},
				MinModes: 1,
				MaxModes: 2,
				// With two modes, choosing both pays the entwine cost once.
				EscalateCost: opt.Val(cost.Mana{cost.O(3), cost.R}),
			}),
		},
	}
}

func resolveRainOfRustArtifact(ctx cardprogram.Context) error {
	_, err := ctx.Destroy(game.Destroy{Object: game.TargetPermanentReference(0)})
	return err
}

func resolveRainOfRustLand(ctx cardprogram.Context) error {
	_, err := ctx.Destroy(game.Destroy{Object: game.TargetPermanentReference(0)})
	return err
}
