package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PlanarGate is the card definition for Planar Gate.
//
// Type: Artifact
// Cost: {6}
//
// Oracle text:
//
//	Creature spells you cast cost {2} less to cast.
var PlanarGate = newPlanarGate

func newPlanarGate() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Planar Gate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creature spells you cast cost {2} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 2,
								CardSelection:    game.Selection{RequiredTypes: []types.Card{types.Creature}},
							},
						},
					},
				},
			},
			OracleText: `
			Creature spells you cast cost {2} less to cast.
		`,
		},
	}
}
