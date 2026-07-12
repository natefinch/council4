package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LethargyTrap is the card definition for Lethargy Trap.
//
// Type: Instant — Trap
// Cost: {3}{U}
//
// Oracle text:
//
//	If three or more creatures are attacking, you may pay {U} rather than pay this spell's mana cost.
//	Attacking creatures get -3/-0 until end of turn.
var LethargyTrap = newLethargyTrap

func newLethargyTrap() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Lethargy Trap",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Instant},
			Subtypes: []types.Sub{types.Trap},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:          "Pay {U}",
					ManaCost:       opt.Val(cost.Mana{cost.U}),
					Condition:      cost.AlternativeConditionCreaturesAttacking,
					ConditionCount: 3,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:      game.LayerPowerToughnessModify,
									Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),
									PowerDelta: -3,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			If three or more creatures are attacking, you may pay {U} rather than pay this spell's mana cost.
			Attacking creatures get -3/-0 until end of turn.
		`,
		},
	}
}
