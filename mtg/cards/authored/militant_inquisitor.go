package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const militantInquisitorOracleID = "d1e326e8-6e9e-4dbd-8c2f-0bef90db2bb7"
const militantInquisitorOracleText = "This creature gets +1/+0 for each Equipment you control."

var MilitantInquisitor = cardset.Entry{
	OracleID: militantInquisitorOracleID,
	Name:     "Militant Inquisitor",
	New:      newMilitantInquisitor,
	Verification: "mtg/rules/authored_existing_paragraph_cards_test.go: " +
		"TestMilitantInquisitorScenario",
}

func newMilitantInquisitor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:     "Militant Inquisitor",
			ManaCost: opt.Val(cost.Mana{cost.O(2), cost.W}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Human, types.Cleric},
			Power:    opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{
				Value: 3,
			}),
			StaticAbilities: []game.StaticAbility{{
				Text: militantInquisitorOracleText,
				ContinuousEffects: []game.ContinuousEffect{{
					Duration:       game.DurationPermanent,
					AffectedSource: true,
					Layer:          game.LayerPowerToughnessModify,
					PowerDeltaDynamic: opt.Val(game.DynamicAmount{
						Kind:       game.DynamicAmountCountSelector,
						Multiplier: 1,
						Group: game.BattlefieldGroup(game.Selection{
							RequiredTypes: []types.Card{types.Artifact},
							SubtypesAny:   []types.Sub{types.Equipment},
							Controller:    game.ControllerYou,
						}),
					}),
				}},
			}},
			OracleText: militantInquisitorOracleText,
		},
	}
}
