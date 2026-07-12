package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NeedlebiteTrap is the card definition for Needlebite Trap.
//
// Type: Instant — Trap
// Cost: {5}{B}{B}
//
// Oracle text:
//
//	If an opponent gained life this turn, you may pay {B} rather than pay this spell's mana cost.
//	Target player loses 5 life and you gain 5 life.
var NeedlebiteTrap = newNeedlebiteTrap

func newNeedlebiteTrap() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Needlebite Trap",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.B,
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
			Types:    []types.Card{types.Instant},
			Subtypes: []types.Sub{types.Trap},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:     "Pay {B}",
					ManaCost:  opt.Val(cost.Mana{cost.B}),
					Condition: cost.AlternativeConditionOpponentGainedLifeThisTurn,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.LoseLife{
							Amount: game.Fixed(5),
							Player: game.TargetPlayerReference(0),
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(5),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			If an opponent gained life this turn, you may pay {B} rather than pay this spell's mana cost.
			Target player loses 5 life and you gain 5 life.
		`,
		},
	}
}
