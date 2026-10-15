package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JetMedallion is the card definition for Jet Medallion.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	Black spells you cast cost {1} less to cast.
var JetMedallion = newJetMedallion

func newJetMedallion() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Jet Medallion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Black spells you cast cost {1} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 1,
								CardSelection:    game.Selection{ColorsAny: []color.Color{color.Black}},
							},
						},
					},
				},
			},
			OracleText: `
			Black spells you cast cost {1} less to cast.
		`,
		},
	}
}
