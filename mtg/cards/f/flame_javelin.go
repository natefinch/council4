package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FlameJavelin is the card definition for Flame Javelin.
//
// Type: Instant
// Cost: {2/R}{2/R}{2/R}
//
// Oracle text:
//
//	({2/R} can be paid with any two mana or with {R}. This card's mana value is 6.)
//	Flame Javelin deals 4 damage to any target.
var FlameJavelin = newFlameJavelin

func newFlameJavelin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Flame Javelin",
			ManaCost: opt.Val(cost.Mana{
				cost.Twobrid(mana.R),
				cost.Twobrid(mana.R),
				cost.Twobrid(mana.R),
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
							Amount:    game.Fixed(4),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			({2/R} can be paid with any two mana or with {R}. This card's mana value is 6.)
			Flame Javelin deals 4 damage to any target.
		`,
		},
	}
}
