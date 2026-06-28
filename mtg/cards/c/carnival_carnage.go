package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Carnival is the card definition for Carnival // Carnage.
//
// Type: Instant // Sorcery
// Cost: {B/R} // {2}{B}{R}
// Face: Carnage — Sorcery ({2}{B}{R})
//
// Oracle text:
//
//	Carnival deals 1 damage to target creature or planeswalker and 1 damage to that permanent's controller.
var Carnival = newCarnival()

func newCarnival() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Carnival",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.B, mana.R),
			}),
			Colors: []color.Color{color.Black, color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature or planeswalker",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(1),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(1),
							Recipient: game.PlayerDamageRecipient(game.ObjectControllerReference(game.TargetPermanentReference(0))),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Carnival deals 1 damage to target creature or planeswalker and 1 damage to that permanent's controller.
		`,
		},
		Layout: game.LayoutSplit,
		Alternate: opt.Val(game.CardFace{
			Name: "Carnage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.R,
			}),
			Colors: []color.Color{color.Black, color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target opponent",
						Allow:      game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
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
						Primitive: game.Discard{
							Amount: game.Fixed(2),
							Player: game.TargetPlayerReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Carnage deals 3 damage to target opponent. That player discards two cards.
		`,
		}),
	}
}
