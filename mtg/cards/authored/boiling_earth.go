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

const boilingEarthOracleText = "Boiling Earth deals 1 damage to each creature your opponents control.\nAwaken 4—{6}{R} (If you cast this spell for {6}{R}, also put four +1/+1 counters on target land you control and it becomes a 0/0 Elemental creature with haste. It's still a land.)"
const boilingEarthAwakenCost cost.AlternativeIndex = 1

var BoilingEarthProgram = game.ProgramReference{
	ID:      "fb34b671-b61a-47e2-90fa-dbe5cf6d3743/spell/0",
	Version: "1",
}

var BoilingEarth = cardset.Entry{
	OracleID:     "fb34b671-b61a-47e2-90fa-dbe5cf6d3743",
	Name:         "Boiling Earth",
	New:          newBoilingEarth,
	Programs:     []cardprogram.Program{{Reference: BoilingEarthProgram, Resolve: resolveBoilingEarth}},
	Verification: "mtg/rules/boiling_earth_test.go: TestBoilingEarthScenario",
}

func newBoilingEarth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Boiling Earth",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.R}),
			Types:      []types.Card{types.Sorcery},
			Colors:     []color.Color{color.Red},
			OracleText: boilingEarthOracleText,
			AlternativeCosts: []cost.Alternative{{
				Label:    "Awaken",
				ManaCost: opt.Val(cost.Mana{cost.O(6), cost.R}),
			}},
			SpellAbility: opt.Val(game.Mode{
				Text:    boilingEarthOracleText,
				Program: BoilingEarthProgram,
				Targets: []game.TargetSpec{{
					MinTargets:      1,
					MaxTargets:      1,
					Constraint:      "target land you control",
					Allow:           game.TargetAllowPermanent,
					AlternativeCost: boilingEarthAwakenCost,
					Selection: opt.Val(game.Selection{
						RequiredTypes: []types.Card{types.Land},
						Controller:    game.ControllerYou,
					}),
				}},
			}.Ability()),
		},
	}
}

func resolveBoilingEarth(ctx cardprogram.Context) error {
	if _, err := ctx.Execute(game.Damage{
		Amount: game.Fixed(1),
		Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Controller:    game.ControllerOpponent,
		})),
	}); err != nil {
		return err
	}
	if ctx.View().Source().AlternativeCost != boilingEarthAwakenCost {
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
