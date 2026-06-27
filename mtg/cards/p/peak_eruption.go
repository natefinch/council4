package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PeakEruption is the card definition for Peak Eruption.
//
// Type: Sorcery
// Cost: {2}{R}
//
// Oracle text:
//
//	Destroy target Mountain. Peak Eruption deals 3 damage to that land's controller.
var PeakEruption = newPeakEruption()

func newPeakEruption() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Peak Eruption",
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
						Constraint: "target Mountain",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Mountain")}}),
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
			Destroy target Mountain. Peak Eruption deals 3 damage to that land's controller.
		`,
		},
	}
}
