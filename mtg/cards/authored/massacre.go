package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var Massacre = cardset.Entry{
	OracleID:     "ce1866e9-28de-4bcc-9999-11958e6c13a6",
	Name:         "Massacre",
	New:          newMassacre,
	Verification: "mtg/rules/massacre_test.go: TestAuthoredMassacreScenario",
}

func newMassacre() *game.CardDef {
	const effectText = "All creatures get -2/-2 until end of turn."

	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:       "Massacre",
			OracleText: "If an opponent controls a Plains and you control a Swamp, you may cast this spell without paying its mana cost.\n" + effectText,
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.B, cost.B}),
			Colors:     []color.Color{color.Black},
			Types:      []types.Card{types.Sorcery},
			AlternativeCosts: []cost.Alternative{{
				Label:    "Cast without paying its mana cost",
				ManaCost: opt.Val(cost.Mana{}),
				Conditions: []cost.AlternativePredicate{
					{Kind: cost.AlternativeConditionOpponentControlsPermanentSubtype, Subtype: types.Plains},
					{Kind: cost.AlternativeConditionControlsPermanentSubtype, Subtype: types.Swamp},
				},
			}},
			SpellAbility: opt.Val(game.Mode{
				Text: effectText,
				Sequence: []game.Instruction{{
					Primitive: game.ApplyContinuous{
						ContinuousEffects: []game.ContinuousEffect{{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							PowerDelta:     -2,
							ToughnessDelta: -2,
						}},
						Duration: game.DurationUntilEndOfTurn,
					},
				}},
			}.Ability()),
		},
	}
}
