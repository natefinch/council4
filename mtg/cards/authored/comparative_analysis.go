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

const comparativeAnalysisOracleID = "59b869b9-cb48-4152-ad91-52f582e2df5b"

var ComparativeAnalysisProgram = game.ProgramReference{
	ID: comparativeAnalysisOracleID + "/spell/0", Version: "1",
}

var ComparativeAnalysis = cardset.Entry{
	OracleID: comparativeAnalysisOracleID,
	Name:     "Comparative Analysis",
	New:      newComparativeAnalysis,
	Programs: []cardprogram.Program{{
		Reference: ComparativeAnalysisProgram,
		Resolve:   resolveComparativeAnalysis,
	}},
	Verification: "mtg/rules/comparative_analysis_test.go: " +
		"TestComparativeAnalysisSurgeDrawsTargetPlayerTwoCards",
}

func newComparativeAnalysis() *game.CardDef {
	const text = "Surge {2}{U} (You may cast this spell for its surge cost if you or a teammate has cast another spell this turn.)\nTarget player draws two cards."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Comparative Analysis",
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.U}),
			Colors:     []color.Color{color.Blue},
			Types:      []types.Card{types.Instant},
			OracleText: text,
			AlternativeCosts: []cost.Alternative{{
				Label:     "Surge {2}{U}",
				ManaCost:  opt.Val(cost.Mana{cost.O(2), cost.U}),
				Condition: cost.AlternativeConditionYourTeamCastSpellThisTurn,
			}},
			SpellAbility: opt.Val(game.Mode{
				Text: "Target player draws two cards.",
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target player",
					Allow:      game.TargetAllowPlayer,
				}},
				Program: ComparativeAnalysisProgram,
			}.Ability()),
		},
	}
}

func resolveComparativeAnalysis(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.Draw{
		Player: game.TargetPlayerReference(0),
		Amount: game.Fixed(2),
	})
	return err
}
