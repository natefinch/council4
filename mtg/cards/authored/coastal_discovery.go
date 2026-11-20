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

const coastalDiscoveryText = "Draw two cards.\nAwaken 4—{5}{U} (If you cast this spell for {5}{U}, also put four +1/+1 counters on target land you control and it becomes a 0/0 Elemental creature with haste. It's still a land.)"

const coastalDiscoveryAwakenCost cost.AlternativeIndex = 1

var CoastalDiscoveryProgram = game.ProgramReference{
	ID:      "f7c84690-8c7c-41a6-b430-a9771030f903/spell/0",
	Version: "1",
}

var CoastalDiscovery = cardset.Entry{
	OracleID: "f7c84690-8c7c-41a6-b430-a9771030f903",
	Name:     "Coastal Discovery",
	New:      newCoastalDiscovery,
	Programs: []cardprogram.Program{{Reference: CoastalDiscoveryProgram, Resolve: resolveCoastalDiscovery}},
	Verification: "mtg/rules/coastal_discovery_test.go: " +
		"TestCoastalDiscoveryAwakenAlternativeDrawsAndAnimatesForest; " +
		"TestCoastalDiscoveryBaseCostDrawsWithoutTarget",
}

func newCoastalDiscovery() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Coastal Discovery",
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.U}),
			Colors:     []color.Color{color.Blue},
			Types:      []types.Card{types.Sorcery},
			OracleText: coastalDiscoveryText,
			AlternativeCosts: []cost.Alternative{{
				Label:    "Awaken",
				ManaCost: opt.Val(cost.Mana{cost.O(5), cost.U}),
			}},
			SpellAbility: opt.Val(game.Mode{
				Text:    coastalDiscoveryText,
				Program: CoastalDiscoveryProgram,
				Targets: []game.TargetSpec{{
					MinTargets:      1,
					MaxTargets:      1,
					Constraint:      "target land you control",
					Allow:           game.TargetAllowPermanent,
					AlternativeCost: coastalDiscoveryAwakenCost,
					Selection: opt.Val(game.Selection{
						RequiredTypes: []types.Card{types.Land},
						Controller:    game.ControllerYou,
					}),
				}},
			}.Ability()),
		},
	}
}

func resolveCoastalDiscovery(ctx cardprogram.Context) error {
	if _, err := ctx.Execute(game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(2)}); err != nil {
		return err
	}
	if ctx.View().Source().AlternativeCost != coastalDiscoveryAwakenCost {
		return nil
	}
	if _, err := ctx.Execute(game.AddCounter{
		Object:      game.TargetPermanentReference(0),
		Amount:      game.Fixed(4),
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
