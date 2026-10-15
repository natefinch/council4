package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JayaSImmolatingInferno is the card definition for Jaya's Immolating Inferno.
//
// Type: Legendary Sorcery
// Cost: {X}{R}{R}
//
// Oracle text:
//
//	(You may cast a legendary sorcery only if you control a legendary creature or planeswalker.)
//	Jaya's Immolating Inferno deals X damage to each of up to three targets.
var JayaSImmolatingInferno = newJayaSImmolatingInferno

func newJayaSImmolatingInferno() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Jaya's Immolating Inferno",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.R,
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 3,
						Constraint: "up to three targets",
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
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Recipient: game.AnyTargetDamageRecipient(1),
						},
					},
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Recipient: game.AnyTargetDamageRecipient(2),
						},
					},
				},
			}.Ability()),
			OracleText: `
			(You may cast a legendary sorcery only if you control a legendary creature or planeswalker.)
			Jaya's Immolating Inferno deals X damage to each of up to three targets.
		`,
		},
	}
}
