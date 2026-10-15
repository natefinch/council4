package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UrzaSIncubator is the card definition for Urza's Incubator.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	As this artifact enters, choose a creature type.
//	Creature spells of the chosen type cost {2} less to cast.
var UrzaSIncubator = newUrzaSIncubator

func newUrzaSIncubator() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Urza's Incubator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creature spells of the chosen type cost {2} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:                         game.CostModifierSpell,
								GenericReduction:             2,
								ChosenSubtypeFromEntryChoice: true,
								CardSelection:                game.Selection{RequiredTypes: []types.Card{types.Creature}},
							},
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryTypeChoiceReplacement("As this artifact enters, choose a creature type."),
			},
			OracleText: `
			As this artifact enters, choose a creature type.
			Creature spells of the chosen type cost {2} less to cast.
		`,
		},
	}
}
