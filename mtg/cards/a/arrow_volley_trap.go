package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArrowVolleyTrap is the card definition for Arrow Volley Trap.
//
// Type: Instant — Trap
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	If four or more creatures are attacking, you may pay {1}{W} rather than pay this spell's mana cost.
//	Arrow Volley Trap deals 5 damage divided as you choose among any number of target attacking creatures.
var ArrowVolleyTrap = newArrowVolleyTrap

func newArrowVolleyTrap() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Arrow Volley Trap",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Instant},
			Subtypes: []types.Sub{types.Trap},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:          "Pay {1}{W}",
					ManaCost:       opt.Val(cost.Mana{cost.O(1), cost.W}),
					Condition:      cost.AlternativeConditionCreaturesAttacking,
					ConditionCount: 4,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 5,
						Constraint: "any number of target attacking creatures",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(5),
							Recipient: game.AnyTargetDamageRecipient(0),
							Divided:   true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			If four or more creatures are attacking, you may pay {1}{W} rather than pay this spell's mana cost.
			Arrow Volley Trap deals 5 damage divided as you choose among any number of target attacking creatures.
		`,
		},
	}
}
