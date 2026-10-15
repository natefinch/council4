package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MjölnirSMight is the card definition for Mjölnir's Might.
//
// Type: Sorcery
// Cost: {3}{R}
//
// Oracle text:
//
//	Mjölnir's Might deals 4 damage to target player.
//	Exile the top card of your library. Until the end of your next turn, you may play that card.
var MjölnirSMight = newMjölnirSMight

func newMjölnirSMight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Mjölnir's Might",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(4),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
					{
						Primitive: game.ImpulseExile{
							Player:   game.ControllerReference(),
							Amount:   game.Fixed(1),
							Duration: game.DurationUntilEndOfYourNextTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Mjölnir's Might deals 4 damage to target player.
			Exile the top card of your library. Until the end of your next turn, you may play that card.
		`,
		},
	}
}
