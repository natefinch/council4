package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Electrodominance is the card definition for Electrodominance.
//
// Type: Instant
// Cost: {X}{R}{R}
//
// Oracle text:
//
//	Electrodominance deals X damage to any target. You may cast a spell with mana value X or less from your hand without paying its mana cost.
var Electrodominance = newElectrodominance

func newElectrodominance() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Electrodominance",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
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
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
					{
						Primitive: game.CastForFree{
							Player:            game.ControllerReference(),
							Selection:         game.Selection{ExcludedTypes: []types.Card{types.Land}},
							Zone:              zone.Hand,
							MaxManaValueFromX: true,
						},
						Optional: true,
					},
				},
			}.Ability()),
			OracleText: `
			Electrodominance deals X damage to any target. You may cast a spell with mana value X or less from your hand without paying its mana cost.
		`,
		},
	}
}
