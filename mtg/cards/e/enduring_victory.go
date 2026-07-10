package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EnduringVictory is the card definition for Enduring Victory.
//
// Type: Instant
// Cost: {4}{W}
//
// Oracle text:
//
//	Destroy target attacking or blocking creature. Bolster 1. (Choose a creature with the least toughness among creatures you control and put a +1/+1 counter on it.)
var EnduringVictory = newEnduringVictory

func newEnduringVictory() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Enduring Victory",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target attacking or blocking creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, CombatState: game.CombatStateAttackingOrBlocking}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Bolster{
							Amount: game.Fixed(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target attacking or blocking creature. Bolster 1. (Choose a creature with the least toughness among creatures you control and put a +1/+1 counter on it.)
		`,
		},
	}
}
