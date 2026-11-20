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

const rushOfIceOracleID = "e9d62416-c4b1-4982-9593-174d46e975c7"
const rushOfIceText = "Tap target creature. It doesn't untap during its controller's next untap step.\nAwaken 3—{4}{U} (If you cast this spell for {4}{U}, also put three +1/+1 counters on target land you control and it becomes a 0/0 Elemental creature with haste. It's still a land.)"
const rushOfIceAwakenCost cost.AlternativeIndex = 1

var RushOfIceProgram = game.ProgramReference{
	ID:      rushOfIceOracleID + "/spell/0",
	Version: "1",
}

var RushOfIce = cardset.Entry{
	OracleID: rushOfIceOracleID,
	Name:     "Rush of Ice",
	New:      newRushOfIce,
	Programs: []cardprogram.Program{{Reference: RushOfIceProgram, Resolve: resolveRushOfIce}},
	Verification: "mtg/rules/rush_of_ice_test.go: " +
		"TestRushOfIceScenario",
}

func newRushOfIce() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Rush of Ice",
			ManaCost:   opt.Val(cost.Mana{cost.U}),
			Colors:     []color.Color{color.Blue},
			Types:      []types.Card{types.Sorcery},
			OracleText: rushOfIceText,
			AlternativeCosts: []cost.Alternative{{
				Label:    "Awaken",
				ManaCost: opt.Val(cost.Mana{cost.O(4), cost.U}),
			}},
			SpellAbility: opt.Val(game.Mode{
				Text:    rushOfIceText,
				Program: RushOfIceProgram,
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
						AlternativeCost: rushOfIceAwakenCost,
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

func resolveRushOfIce(ctx cardprogram.Context) error {
	if _, err := ctx.Execute(game.Tap{Object: game.TargetPermanentReference(0)}); err != nil {
		return err
	}
	if _, err := ctx.Execute(game.SkipNextUntap{Object: game.TargetPermanentReference(0)}); err != nil {
		return err
	}
	if ctx.View().Source().AlternativeCost != rushOfIceAwakenCost {
		return nil
	}
	if _, err := ctx.Execute(game.AddCounter{
		Object:      game.TargetPermanentReference(1),
		Amount:      game.Fixed(3),
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
