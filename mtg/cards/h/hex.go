package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Hex is the card definition for Hex.
//
// Type: Sorcery
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	Destroy six target creatures.
var Hex = newHex

func newHex() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Hex",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 6,
						MaxTargets: 6,
						Constraint: "six target creatures",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(1),
						},
					},
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(2),
						},
					},
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(3),
						},
					},
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(4),
						},
					},
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(5),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy six target creatures.
		`,
		},
	}
}
