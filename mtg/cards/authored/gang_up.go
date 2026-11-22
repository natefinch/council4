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

var gangUpProgram = game.ProgramReference{
	ID:      "dea311c3-9be3-4059-a206-5fbbdca4d6c9/spell/0",
	Version: "1",
}

var GangUp = cardset.Entry{
	OracleID: "dea311c3-9be3-4059-a206-5fbbdca4d6c9",
	Name:     "Gang Up",
	New:      newGangUp,
	Programs: []cardprogram.Program{{Reference: gangUpProgram, Resolve: resolveGangUp}},
	Verification: "mtg/rules/gang_up_test.go: TestGangUpAssistXTargetCreatureAndDestroys; " +
		"CR 2026-11-06 107.3a/i, 115.1a, 208.1, 601.2b/c/f-h, 608.2b, " +
		"613.4a-d/5, 701.8a-c, 702.132a",
}

func newGangUp() *game.CardDef {
	const text = "Assist (Another player can pay up to {X} of this spell's cost. You choose the value of X.)\n" +
		"Destroy target creature with power X or less."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:            "Gang Up",
			ManaCost:        opt.Val(cost.Mana{cost.X, cost.B}),
			Types:           []types.Card{types.Instant},
			Colors:          []color.Color{color.Black},
			OracleText:      text,
			StaticAbilities: []game.StaticAbility{game.AssistStaticBody},
			SpellAbility: opt.Val(game.Mode{
				Text:    "Destroy target creature with power X or less.",
				Program: gangUpProgram,
				Targets: []game.TargetSpec{{
					MinTargets:   1,
					MaxTargets:   1,
					Constraint:   "target creature with power X or less",
					Allow:        game.TargetAllowPermanent,
					Selection:    opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
					PowerAtMostX: true,
				}},
			}.Ability()),
		},
	}
}

func resolveGangUp(ctx cardprogram.Context) error {
	_, err := ctx.Destroy(game.Destroy{Object: game.TargetPermanentReference(0)})
	return err
}
