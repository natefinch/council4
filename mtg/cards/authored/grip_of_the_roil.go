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

const (
	gripOfTheRoilOracleID = "c3c87292-2fc8-4759-b7b6-18a8b539b5d1"
	gripOfTheRoilText     = "Surge {1}{U} (You may cast this spell for its surge cost if you or a teammate has cast another spell this turn.)\nTap target creature. It doesn't untap during its controller's next untap step.\nDraw a card."
)

var GripOfTheRoilProgram = game.ProgramReference{
	ID:      gripOfTheRoilOracleID + "/spell/0",
	Version: "1",
}

var GripOfTheRoil = cardset.Entry{
	OracleID: gripOfTheRoilOracleID,
	Name:     "Grip of the Roil",
	New:      newGripOfTheRoil,
	Programs: []cardprogram.Program{{Reference: GripOfTheRoilProgram, Resolve: resolveGripOfTheRoil}},
	Verification: "mtg/rules/grip_of_the_roil_test.go: " +
		"TestGripOfTheRoilScenario",
}

func newGripOfTheRoil() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Grip of the Roil",
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.U}),
			Colors:     []color.Color{color.Blue},
			Types:      []types.Card{types.Instant},
			OracleText: gripOfTheRoilText,
			AlternativeCosts: []cost.Alternative{{
				Label:     "Surge {1}{U}",
				ManaCost:  opt.Val(cost.Mana{cost.O(1), cost.U}),
				Condition: cost.AlternativeConditionYourTeamCastSpellThisTurn,
			}},
			SpellAbility: opt.Val(game.Mode{
				Text:    "Tap target creature. It doesn't untap during its controller's next untap step.\nDraw a card.",
				Program: GripOfTheRoilProgram,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature",
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypes: []types.Card{types.Creature},
					}),
				}},
			}.Ability()),
		},
	}
}

func resolveGripOfTheRoil(ctx cardprogram.Context) error {
	if _, err := ctx.Execute(game.Tap{Object: game.TargetPermanentReference(0)}); err != nil {
		return err
	}
	if _, err := ctx.Execute(game.SkipNextUntap{Object: game.TargetPermanentReference(0)}); err != nil {
		return err
	}
	_, err := ctx.Execute(game.Draw{
		Player: game.ControllerReference(),
		Amount: game.Fixed(1),
	})
	return err
}
