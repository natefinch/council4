package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SphereOfResistance is the card definition for Sphere of Resistance.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	Spells cost {1} more to cast.
var SphereOfResistance = newSphereOfResistance

func newSphereOfResistance() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Sphere of Resistance",
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
								Kind:            game.CostModifierSpell,
								GenericIncrease: 1,
							},
						},
					},
				},
			},
			OracleText: `
			Spells cost {1} more to cast.
		`,
		},
	}
}
