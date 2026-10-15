package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WarHorn is the card definition for War Horn.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	Attacking creatures you control get +1/+0.
var WarHorn = newWarHorn

func newWarHorn() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "War Horn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Attacking creatures you control get +1/+0.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 1,
						},
					},
				},
			},
			OracleText: `
			Attacking creatures you control get +1/+0.
		`,
		},
	}
}
