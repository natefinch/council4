package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HanabiBlast is the card definition for Hanabi Blast.
//
// Type: Instant
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	Hanabi Blast deals 2 damage to any target. Return Hanabi Blast to its owner's hand, then discard a card at random.
var HanabiBlast = newHanabiBlast

func newHanabiBlast() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Hanabi Blast",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
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
						Primitive: game.Damage{
							Amount:    game.Fixed(2),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
					{
						Primitive: game.MoveResolvingSpell{
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.Discard{
							Amount:   game.Fixed(1),
							Player:   game.ControllerReference(),
							AtRandom: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Hanabi Blast deals 2 damage to any target. Return Hanabi Blast to its owner's hand, then discard a card at random.
		`,
		},
	}
}
