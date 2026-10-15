package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SoothingOfSméagol is the card definition for Soothing of Sméagol.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Return target nontoken creature to its owner's hand. The Ring tempts you.
var SoothingOfSméagol = newSoothingOfSméagol

func newSoothingOfSméagol() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Soothing of Sméagol",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target nontoken creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, NonToken: true}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(0),
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.RingTempts{
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return target nontoken creature to its owner's hand. The Ring tempts you.
		`,
		},
	}
}
