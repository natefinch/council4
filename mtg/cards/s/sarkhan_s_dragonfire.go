package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SarkhanSDragonfire is the card definition for Sarkhan's Dragonfire.
//
// Type: Sorcery
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Sarkhan's Dragonfire deals 3 damage to any target.
//	Look at the top five cards of your library. You may reveal a red card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
var SarkhanSDragonfire = newSarkhanSDragonfire()

func newSarkhanSDragonfire() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Sarkhan's Dragonfire",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "any target",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(3),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
					{
						Primitive: game.Dig{
							Player:    game.ControllerReference(),
							Look:      game.Fixed(5),
							Take:      game.Fixed(1),
							Remainder: game.DigRemainderLibraryBottom,
							Filter:    opt.Val(game.Selection{ColorsAny: []color.Color{color.Red}}),
							TakeUpTo:  true,
							Reveal:    true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Sarkhan's Dragonfire deals 3 damage to any target.
			Look at the top five cards of your library. You may reveal a red card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
		`,
		},
	}
}
