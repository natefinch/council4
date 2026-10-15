package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TelJiladJustice is the card definition for Tel-Jilad Justice.
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	Destroy target artifact. Scry 2. (Look at the top two cards of your library, then put any number of them on the bottom and the rest on top in any order.)
var TelJiladJustice = newTelJiladJustice

func newTelJiladJustice() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tel-Jilad Justice",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target artifact",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Scry{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target artifact. Scry 2. (Look at the top two cards of your library, then put any number of them on the bottom and the rest on top in any order.)
		`,
		},
	}
}
