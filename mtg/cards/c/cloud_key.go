package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CloudKey is the card definition for Cloud Key.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	As this artifact enters, choose artifact, creature, enchantment, instant, or sorcery.
//	Spells you cast of the chosen type cost {1} less to cast.
var CloudKey = newCloudKey

func newCloudKey() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Cloud Key",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Spells you cast of the chosen type cost {1} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:                          game.CostModifierSpell,
								GenericReduction:              1,
								ChosenCardTypeFromEntryChoice: true,
							},
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryCardTypeChoiceReplacement("As this artifact enters, choose artifact, creature, enchantment, instant, or sorcery."),
			},
			OracleText: `
			As this artifact enters, choose artifact, creature, enchantment, instant, or sorcery.
			Spells you cast of the chosen type cost {1} less to cast.
		`,
		},
	}
}
