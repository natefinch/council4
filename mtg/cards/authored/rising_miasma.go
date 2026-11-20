package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const risingMiasmaOracleID = "83fb7dfb-64a4-4320-b544-830c29f24df9"
const risingMiasmaText = "All creatures get -2/-2 until end of turn.\nAwaken 3—{5}{B}{B} (If you cast this spell for {5}{B}{B}, also put three +1/+1 counters on target land you control and it becomes a 0/0 Elemental creature with haste. It's still a land.)"
const risingMiasmaAwakenCost cost.AlternativeIndex = 1

var RisingMiasmaProgram = game.ProgramReference{
	ID:      risingMiasmaOracleID + "/spell/0",
	Version: "1",
}

var RisingMiasma = cardset.Entry{
	OracleID: risingMiasmaOracleID,
	Name:     "Rising Miasma",
	New:      newRisingMiasma,
	Programs: []cardprogram.Program{{Reference: RisingMiasmaProgram, Resolve: resolveRisingMiasma}},
	Verification: "mtg/rules/rising_miasma_test.go: " +
		"TestRisingMiasmaScenario",
}

func newRisingMiasma() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:       "Rising Miasma",
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.B}),
			Colors:     []color.Color{color.Black},
			Types:      []types.Card{types.Sorcery},
			OracleText: risingMiasmaText,
			AlternativeCosts: []cost.Alternative{{
				Label:    "Awaken",
				ManaCost: opt.Val(cost.Mana{cost.O(5), cost.B, cost.B}),
			}},
			SpellAbility: opt.Val(game.Mode{
				Text:    risingMiasmaText,
				Program: RisingMiasmaProgram,
				Targets: []game.TargetSpec{{
					MinTargets:      1,
					MaxTargets:      1,
					Constraint:      "target land you control",
					Allow:           game.TargetAllowPermanent,
					AlternativeCost: risingMiasmaAwakenCost,
					Selection: opt.Val(game.Selection{
						RequiredTypes: []types.Card{types.Land},
						Controller:    game.ControllerYou,
					}),
				}},
			}.Ability()),
		},
	}
}

func resolveRisingMiasma(ctx cardprogram.Context) error {
	if _, err := ctx.Execute(game.ApplyContinuous{
		ContinuousEffects: []game.ContinuousEffect{{
			Layer:          game.LayerPowerToughnessModify,
			Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
			PowerDelta:     -2,
			ToughnessDelta: -2,
		}},
		Duration: game.DurationUntilEndOfTurn,
	}); err != nil {
		return err
	}
	if ctx.View().Source().AlternativeCost != risingMiasmaAwakenCost {
		return nil
	}
	if _, err := ctx.Execute(game.AddCounter{
		Object:      game.TargetPermanentReference(0),
		Amount:      game.Fixed(3),
		CounterKind: counter.PlusOnePlusOne,
	}); err != nil {
		return err
	}
	_, err := ctx.Execute(game.ApplyContinuous{
		Object: opt.Val(game.TargetPermanentReference(0)),
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
