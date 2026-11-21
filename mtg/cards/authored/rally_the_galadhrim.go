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

var rallyTheGaladhrimProgram = game.ProgramReference{
	ID: "84b2d0f6-ad98-4a71-a3d5-645b0f2db6c6/spell/0", Version: "1",
}

var RallyTheGaladhrim = cardset.Entry{
	OracleID: "84b2d0f6-ad98-4a71-a3d5-645b0f2db6c6",
	Name:     "Rally the Galadhrim",
	New:      newRallyTheGaladhrim,
	Programs: []cardprogram.Program{{
		Reference: rallyTheGaladhrimProgram, Resolve: resolveRallyTheGaladhrim,
	}},
	Verification: "mtg/rules/rally_the_galadhrim_test.go: TestRallyTheGaladhrimCreatesTargetedCreatureCopiesWithConspire",
}

const rallyTheGaladhrimText = "Create a token that's a copy of target creature you control."

func newRallyTheGaladhrim() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green, color.Blue),
		CardFace: game.CardFace{
			Name:       "Rally the Galadhrim",
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.G, cost.U}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.Green, color.Blue},
			OracleText: rallyTheGaladhrimText + "\nConspire (As you cast this spell, you may tap two untapped creatures you control that share a color with it. When you do, copy it and you may choose a new target for the copy.)",
			StaticAbilities: []game.StaticAbility{
				game.ConspireStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Text: rallyTheGaladhrimText,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypes: []types.Card{types.Creature},
						Controller:    game.ControllerYou,
					}),
				}},
				Program: rallyTheGaladhrimProgram,
			}.Ability()),
		},
	}
}

func resolveRallyTheGaladhrim(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.CreateToken{
		Amount: game.Fixed(1),
		Source: game.TokenCopyOf(game.TokenCopySpec{
			Source: game.TokenCopySourceObject,
			Object: game.TargetPermanentReference(0),
		}),
	})
	return err
}
