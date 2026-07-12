package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PitfallTrap is the card definition for Pitfall Trap.
//
// Type: Instant — Trap
// Cost: {2}{W}
//
// Oracle text:
//
//	If exactly one creature is attacking, you may pay {W} rather than pay this spell's mana cost.
//	Destroy target attacking creature without flying.
var PitfallTrap = newPitfallTrap

func newPitfallTrap() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Pitfall Trap",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Instant},
			Subtypes: []types.Sub{types.Trap},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:            "Pay {W}",
					ManaCost:         opt.Val(cost.Mana{cost.W}),
					Condition:        cost.AlternativeConditionCreaturesAttacking,
					ConditionCount:   1,
					ConditionExactly: true,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target attacking creature without flying",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking, ExcludedKeyword: game.Flying}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			If exactly one creature is attacking, you may pay {W} rather than pay this spell's mana cost.
			Destroy target attacking creature without flying.
		`,
		},
	}
}
