package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FoundryInspector is the card definition for Foundry Inspector.
//
// Type: Artifact Creature — Construct
// Cost: {3}
//
// Oracle text:
//
//	Artifact spells you cast cost {1} less to cast.
var FoundryInspector = newFoundryInspector

func newFoundryInspector() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Foundry Inspector",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Artifact spells you cast cost {1} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 1,
								CardSelection:    game.Selection{RequiredTypes: []types.Card{types.Artifact}},
							},
						},
					},
				},
			},
			OracleText: `
			Artifact spells you cast cost {1} less to cast.
		`,
		},
	}
}
