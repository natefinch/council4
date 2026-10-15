package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Weakstone is the card definition for Weakstone.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	Attacking creatures get -1/-0.
var Weakstone = newWeakstone

func newWeakstone() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Weakstone",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Attacking creatures get -1/-0.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: -1,
						},
					},
				},
			},
			OracleText: `
			Attacking creatures get -1/-0.
		`,
		},
	}
}
