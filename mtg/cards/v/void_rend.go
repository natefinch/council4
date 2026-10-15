package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VoidRend is the card definition for Void Rend.
//
// Type: Instant
// Cost: {W}{U}{B}
//
// Oracle text:
//
//	This spell can't be countered.
//	Destroy target nonland permanent.
var VoidRend = newVoidRend

func newVoidRend() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Void Rend",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.U,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.Blue, color.White},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target nonland permanent",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			This spell can't be countered.
			Destroy target nonland permanent.
		`,
		},
	}
}
