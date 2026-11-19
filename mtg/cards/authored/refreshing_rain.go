package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var RefreshingRain = cardset.Entry{
	OracleID:     "c6f69168-71c2-41a2-ab95-cfd7baf0ba68",
	Name:         "Refreshing Rain",
	New:          newRefreshingRain,
	Verification: "mtg/rules/refreshing_rain_test.go: TestAuthoredRefreshingRainScenario",
}

func newRefreshingRain() *game.CardDef {
	const oracleText = "If an opponent controls a Swamp and you control a Forest, you may cast this spell without paying its mana cost.\nTarget player gains 6 life."
	const effectText = "Target player gains 6 life."

	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Refreshing Rain",
			OracleText: oracleText,
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.G}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Green},
			AlternativeCosts: []cost.Alternative{{
				Label:    "Cast without paying its mana cost",
				ManaCost: opt.Val(cost.Mana{}),
				Conditions: []cost.AlternativePredicate{
					{Kind: cost.AlternativeConditionControlsPermanentSubtype, Subtype: types.Forest},
					{Kind: cost.AlternativeConditionOpponentControlsPermanentSubtype, Subtype: types.Swamp},
				},
			}},
			SpellAbility: opt.Val(game.Mode{
				Text: effectText,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target player",
					Allow:      game.TargetAllowPlayer,
				}},
				Sequence: []game.Instruction{{
					Primitive: game.GainLife{
						Player: game.TargetPlayerReference(0),
						Amount: game.Fixed(6),
					},
				}},
			}.Ability()),
		},
	}
}
