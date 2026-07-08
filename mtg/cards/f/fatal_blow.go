package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FatalBlow is the card definition for Fatal Blow.
//
// Type: Instant
// Cost: {B}
//
// Oracle text:
//
//	Destroy target creature that was dealt damage this turn. It can't be regenerated.
var FatalBlow = newFatalBlow

func newFatalBlow() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Fatal Blow",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature that was dealt damage this turn",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, DealtDamageThisTurn: true}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object:              game.TargetPermanentReference(0),
							PreventRegeneration: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target creature that was dealt damage this turn. It can't be regenerated.
		`,
		},
	}
}
