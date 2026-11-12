package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const oxiddaFinisherOracleID = "f308a1de-94a6-4f9f-8f46-6f2018959b40"
const oxiddaFinisherOracleText = "Affinity for Equipment (This spell costs {1} less to cast for each Equipment you control.)\nTrample"

var OxiddaFinisher = cardset.Entry{
	OracleID: oxiddaFinisherOracleID,
	Name:     "Oxidda Finisher",
	New:      newOxiddaFinisher,
	Verification: "mtg/rules/authored_existing_paragraph_cards_test.go: " +
		"TestOxiddaFinisherScenario",
}

func newOxiddaFinisher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:     "Oxidda Finisher",
			ManaCost: opt.Val(cost.Mana{cost.O(5), cost.R, cost.R}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Ogre, types.Rebel},
			Power:    opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{
				Value: 5,
			}),
			StaticAbilities: []game.StaticAbility{
				{
					Text: "Affinity for Equipment (This spell costs {1} less to cast for each Equipment you control.)",
					RuleEffects: []game.RuleEffect{{
						Kind:           game.RuleEffectCostModifier,
						AffectedSource: true,
						CostModifier: game.CostModifier{
							Kind:               game.CostModifierSpell,
							PerObjectReduction: 1,
							CountSelection: &game.Selection{
								RequiredTypes: []types.Card{types.Artifact},
								SubtypesAny:   []types.Sub{types.Equipment},
								Controller:    game.ControllerYou,
							},
						},
					}},
				},
				game.TrampleStaticBody,
			},
			OracleText: oxiddaFinisherOracleText,
		},
	}
}
