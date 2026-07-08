package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CatharticPyre is the card definition for Cathartic Pyre.
//
// Type: Instant
// Cost: {1}{R}
//
// Oracle text:
//
//	Choose one —
//	• Cathartic Pyre deals 3 damage to target creature or planeswalker.
//	• Discard up to two cards, then draw that many cards.
var CatharticPyre = newCatharticPyre

func newCatharticPyre() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Cathartic Pyre",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Cathartic Pyre deals 3 damage to target creature or planeswalker.",
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
									Amount:    game.Fixed(3),
									Recipient: game.AnyTargetDamageRecipient(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Discard up to two cards, then draw that many cards.",
						Sequence: []game.Instruction{
							{
								Primitive: game.DiscardThenDraw{
									Player: game.ControllerReference(),
									Max:    2,
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Cathartic Pyre deals 3 damage to target creature or planeswalker.
			• Discard up to two cards, then draw that many cards.
		`,
		},
	}
}
