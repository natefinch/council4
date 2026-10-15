package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ManaMatrix is the card definition for Mana Matrix.
//
// Type: Artifact
// Cost: {6}
//
// Oracle text:
//
//	Instant and enchantment spells you cast cost {2} less to cast.
var ManaMatrix = newManaMatrix

func newManaMatrix() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Mana Matrix",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Instant and enchantment spells you cast cost {2} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 2,
								CardSelection:    game.Selection{RequiredTypes: []types.Card{types.Instant}},
							},
						},
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 2,
								CardSelection:    game.Selection{RequiredTypes: []types.Card{types.Enchantment}},
							},
						},
					},
				},
			},
			OracleText: `
			Instant and enchantment spells you cast cost {2} less to cast.
		`,
		},
	}
}
