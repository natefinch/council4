package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// WeightAdvantage is the card definition for Weight Advantage.
//
// Type: Conspiracy
//
// Oracle text:
//
//	(Start the game with this conspiracy face up in the command zone.)
//	Each creature you control assigns combat damage equal to its toughness rather than its power.
var WeightAdvantage = newWeightAdvantage()

func newWeightAdvantage() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Weight Advantage",
			Types: []types.Card{types.Conspiracy},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectAssignCombatDamageUsingToughness,
							AffectedController: game.ControllerYou,
							PermanentTypes:     []types.Card{types.Creature},
						},
					},
				},
			},
			OracleText: `
			(Start the game with this conspiracy face up in the command zone.)
			Each creature you control assigns combat damage equal to its toughness rather than its power.
		`,
		},
	}
}
