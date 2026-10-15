package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InspiredUltimatum is the card definition for Inspired Ultimatum.
//
// Type: Sorcery
// Cost: {U}{U}{R}{R}{R}{W}{W}
//
// Oracle text:
//
//	Target player gains 5 life, Inspired Ultimatum deals 5 damage to any target, then you draw five cards.
var InspiredUltimatum = newInspiredUltimatum

func newInspiredUltimatum() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Inspired Ultimatum",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.U,
				cost.R,
				cost.R,
				cost.R,
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.Red, color.Blue, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target player",
						Allow:      game.TargetAllowPlayer,
					},
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "any target",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(5),
							Player: game.TargetPlayerReference(0),
						},
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(5),
							Recipient: game.AnyTargetDamageRecipient(1),
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(5),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player gains 5 life, Inspired Ultimatum deals 5 damage to any target, then you draw five cards.
		`,
		},
	}
}
