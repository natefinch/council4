package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FerozSBan is the card definition for Feroz's Ban.
//
// Type: Artifact
// Cost: {6}
//
// Oracle text:
//
//	Creature spells cost {2} more to cast.
var FerozSBan = newFerozSBan()

func newFerozSBan() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Feroz's Ban",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind: game.RuleEffectCostModifier,
							CostModifier: game.CostModifier{
								Kind:            game.CostModifierSpell,
								CardSelection:   game.Selection{RequiredTypes: []types.Card{types.Creature}},
								GenericIncrease: 2,
							},
						},
					},
				},
			},
			OracleText: `
			Creature spells cost {2} more to cast.
		`,
		},
	}
}
