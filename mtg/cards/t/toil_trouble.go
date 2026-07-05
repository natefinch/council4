package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Toil is the card definition for Toil // Trouble.
//
// Type: Sorcery // Sorcery
// Cost: {2}{B} // {2}{R}
// Face: Trouble — Sorcery ({2}{R})
//
// Oracle text:
//
//	Target player draws two cards and loses 2 life.
//	Fuse (You may cast one or both halves of this card from your hand.)
var Toil = newToil()

func newToil() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Toil",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.FuseStaticBody,
			},
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
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.TargetPlayerReference(0),
						},
					},
					{
						Primitive: game.LoseLife{
							Amount: game.Fixed(2),
							Player: game.TargetPlayerReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player draws two cards and loses 2 life.
			Fuse (You may cast one or both halves of this card from your hand.)
		`,
		},
		Layout: game.LayoutSplit,
		Alternate: opt.Val(game.CardFace{
			Name: "Trouble",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.FuseStaticBody,
			},
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
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountCardsInZone,
								Multiplier: 1,
								Player:     func() *game.PlayerReference { ref := game.EventPlayerReference(); return &ref }(),
								CardZone:   zone.Hand,
								Selection:  &game.Selection{},
							}),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Trouble deals damage to target player equal to the number of cards in that player's hand.
			Fuse (You may cast one or both halves of this card from your hand.)
		`,
		}),
	}
}
