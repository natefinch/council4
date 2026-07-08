package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Cryoclasm is the card definition for Cryoclasm.
//
// Type: Sorcery
// Cost: {2}{R}
//
// Oracle text:
//
//	Destroy target Plains or Island. Cryoclasm deals 3 damage to that land's controller.
var Cryoclasm = newCryoclasm

func newCryoclasm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Cryoclasm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target Plains or Island",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Plains"), types.Sub("Island")}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(3),
							Recipient: game.PlayerDamageRecipient(game.ObjectControllerReference(game.TargetPermanentReference(0))),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target Plains or Island. Cryoclasm deals 3 damage to that land's controller.
		`,
		},
	}
}
