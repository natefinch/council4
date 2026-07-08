package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThornOfAmethyst is the card definition for Thorn of Amethyst.
//
// Type: Artifact
// Cost: {2}
//
// Oracle text:
//
//	Noncreature spells cost {1} more to cast.
var ThornOfAmethyst = newThornOfAmethyst

func newThornOfAmethyst() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Thorn of Amethyst",
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
								CardSelection:   game.Selection{ExcludedTypes: []types.Card{types.Creature}},
								GenericIncrease: 1,
							},
						},
					},
				},
			},
			OracleText: `
			Noncreature spells cost {1} more to cast.
		`,
		},
	}
}
