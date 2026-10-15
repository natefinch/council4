package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RubyMedallion is the card definition for Ruby Medallion.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	Red spells you cast cost {1} less to cast.
var RubyMedallion = newRubyMedallion

func newRubyMedallion() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Ruby Medallion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Red spells you cast cost {1} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 1,
								CardSelection:    game.Selection{ColorsAny: []color.Color{color.Red}},
							},
						},
					},
				},
			},
			OracleText: `
			Red spells you cast cost {1} less to cast.
		`,
		},
	}
}
