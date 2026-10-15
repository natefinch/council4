package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SapphireMedallion is the card definition for Sapphire Medallion.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	Blue spells you cast cost {1} less to cast.
var SapphireMedallion = newSapphireMedallion

func newSapphireMedallion() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Sapphire Medallion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Blue spells you cast cost {1} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 1,
								CardSelection:    game.Selection{ColorsAny: []color.Color{color.Blue}},
							},
						},
					},
				},
			},
			OracleText: `
			Blue spells you cast cost {1} less to cast.
		`,
		},
	}
}
