package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HelmOfAwakening is the card definition for Helm of Awakening.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	Spells cost {1} less to cast.
var HelmOfAwakening = newHelmOfAwakening

func newHelmOfAwakening() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Helm of Awakening",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind: game.RuleEffectCostModifier,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 1,
							},
						},
					},
				},
			},
			OracleText: `
			Spells cost {1} less to cast.
		`,
		},
	}
}
