package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const roilSpoutOracleID = "ebc1d343-26f1-4d85-ac3a-610da24c990e"
const roilSpoutText = "Put target creature on top of its owner's library.\nAwaken 4—{4}{W}{U} (If you cast this spell for {4}{W}{U}, also put four +1/+1 counters on target land you control and it becomes a 0/0 Elemental creature with haste. It's still a land.)"
const roilSpoutAwakenCost cost.AlternativeIndex = 1

var RoilSpoutProgram = game.ProgramReference{
	ID:      roilSpoutOracleID + "/spell/0",
	Version: "1",
}

var RoilSpout = cardset.Entry{
	OracleID: roilSpoutOracleID,
	Name:     "Roil Spout",
	New:      newRoilSpout,
	Programs: []cardprogram.Program{{Reference: RoilSpoutProgram, Resolve: resolveRoilSpout}},
	Verification: "mtg/rules/roil_spout_test.go: " +
		"TestRoilSpoutScenario",
}

func newRoilSpout() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name:       "Roil Spout",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.W, cost.U}),
			Colors:     []color.Color{color.White, color.Blue},
			Types:      []types.Card{types.Sorcery},
			OracleText: roilSpoutText,
			AlternativeCosts: []cost.Alternative{{
				Label:    "Awaken",
				ManaCost: opt.Val(cost.Mana{cost.O(4), cost.W, cost.U}),
			}},
			SpellAbility: opt.Val(game.Mode{
				Text:    roilSpoutText,
				Program: RoilSpoutProgram,
				Targets: []game.TargetSpec{
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{
							RequiredTypes: []types.Card{types.Creature},
						}),
					},
					{
						MinTargets:      1,
						MaxTargets:      1,
						Constraint:      "target land you control",
						Allow:           game.TargetAllowPermanent,
						AlternativeCost: roilSpoutAwakenCost,
						Selection: opt.Val(game.Selection{
							RequiredTypes: []types.Card{types.Land},
							Controller:    game.ControllerYou,
						}),
					},
				},
			}.Ability()),
		},
	}
}

func resolveRoilSpout(ctx cardprogram.Context) error {
	if _, err := ctx.Execute(game.MovePermanent{
		Object:      game.TargetPermanentReference(0),
		Destination: zone.Library,
	}); err != nil {
		return err
	}
	if ctx.View().Source().AlternativeCost != roilSpoutAwakenCost {
		return nil
	}
	if _, err := ctx.Execute(game.AddCounter{
		Object:      game.TargetPermanentReference(1),
		Amount:      game.Fixed(4),
		CounterKind: counter.PlusOnePlusOne,
	}); err != nil {
		return err
	}
	_, err := ctx.Execute(game.ApplyContinuous{
		Object: opt.Val(game.TargetPermanentReference(1)),
		ContinuousEffects: []game.ContinuousEffect{
			{
				Layer:          game.LayerType,
				AddTypes:       []types.Card{types.Creature},
				RemoveSubtypes: types.SubtypesForType(types.Creature),
				AddSubtypes:    []types.Sub{types.Elemental},
			},
			{Layer: game.LayerAbility, AddKeywords: []game.Keyword{game.Haste}},
			{
				Layer:        game.LayerPowerToughnessSet,
				SetPower:     opt.Val(game.PT{Value: 0}),
				SetToughness: opt.Val(game.PT{Value: 0}),
			},
		},
		Duration: game.DurationPermanent,
	})
	return err
}
