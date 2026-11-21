package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var bringDownProgram = game.ProgramReference{
	ID: "05d2b89b-acc4-4289-95a3-a09ebb23a5d6/spell/0", Version: "1",
}

var BringDown = cardset.Entry{
	OracleID: "05d2b89b-acc4-4289-95a3-a09ebb23a5d6",
	Name:     "Bring Down",
	New:      newBringDown,
	Programs: []cardprogram.Program{{Reference: bringDownProgram, Resolve: resolveBringDown}},
	Verification: "mtg/rules/bring_down_assist_test.go: TestBringDownAssistCastAndResolves; " +
		"CR 2026-09-25 115.1a, 208.1, 601.2c/f-h, 605.3a, 608.2b, 613.1g/4c, 701.8a-c, 702.132a",
}

func newBringDown() *game.CardDef {
	const text = "Assist (Another player can pay up to {3} of this spell's cost.)\nDestroy target creature with power 4 or greater."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:            "Bring Down",
			OracleText:      text,
			ManaCost:        opt.Val(cost.Mana{cost.O(3), cost.W}),
			Types:           []types.Card{types.Sorcery},
			Colors:          []color.Color{color.White},
			StaticAbilities: []game.StaticAbility{game.AssistStaticBody},
			SpellAbility: opt.Val(game.Mode{
				Text:    "Destroy target creature with power 4 or greater.",
				Program: bringDownProgram,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypes: []types.Card{types.Creature},
						Power:         opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4}),
					}),
				}},
			}.Ability()),
		},
	}
}

func resolveBringDown(ctx cardprogram.Context) error {
	_, err := ctx.Destroy(game.Destroy{Object: game.TargetPermanentReference(0)})
	return err
}
