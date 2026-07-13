package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GlacialRay is the card definition for Glacial Ray.
//
// Type: Instant — Arcane
// Cost: {1}{R}
//
// Oracle text:
//
//	Glacial Ray deals 2 damage to any target.
//	Splice onto Arcane {1}{R} (As you cast an Arcane spell, you may reveal this card from your hand and pay its splice cost. If you do, add this card's effects to that spell.)
var GlacialRay = newGlacialRay

func newGlacialRay() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Glacial Ray",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Instant},
			Subtypes: []types.Sub{types.Arcane},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.SpliceKeyword{Cost: cost.Mana{cost.O(1), cost.R}},
					},
				},
			},
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
							Amount:    game.Fixed(2),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Glacial Ray deals 2 damage to any target.
			Splice onto Arcane {1}{R} (As you cast an Arcane spell, you may reveal this card from your hand and pay its splice cost. If you do, add this card's effects to that spell.)
		`,
		},
	}
}
