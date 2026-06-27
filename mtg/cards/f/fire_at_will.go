package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FireAtWill is the card definition for Fire at Will.
//
// Type: Instant
// Cost: {R/W}{R/W}{R/W}
//
// Oracle text:
//
//	Fire at Will deals 3 damage divided as you choose among one, two, or three target attacking or blocking creatures.
var FireAtWill = newFireAtWill()

func newFireAtWill() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Fire at Will",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.R, mana.W),
				cost.HybridMana(mana.R, mana.W),
				cost.HybridMana(mana.R, mana.W),
			}),
			Colors: []color.Color{color.Red, color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 3,
						Constraint: "one, two, or three target attacking or blocking creatures",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, CombatState: game.CombatStateAttackingOrBlocking}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(3),
							Recipient: game.AnyTargetDamageRecipient(0),
							Divided:   true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Fire at Will deals 3 damage divided as you choose among one, two, or three target attacking or blocking creatures.
		`,
		},
	}
}
