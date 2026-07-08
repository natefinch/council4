package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Counterintelligence is the card definition for Counterintelligence.
//
// Type: Sorcery
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	Return one or two target creatures to their owners' hands.
var Counterintelligence = newCounterintelligence

func newCounterintelligence() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Counterintelligence",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 2,
						Constraint: "one or two target creatures",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Bounce{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Bounce{
							Object: game.TargetPermanentReference(1),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return one or two target creatures to their owners' hands.
		`,
		},
	}
}
