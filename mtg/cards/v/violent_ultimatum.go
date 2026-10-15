package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ViolentUltimatum is the card definition for Violent Ultimatum.
//
// Type: Sorcery
// Cost: {B}{B}{R}{R}{R}{G}{G}
//
// Oracle text:
//
//	Destroy three target permanents.
var ViolentUltimatum = newViolentUltimatum

func newViolentUltimatum() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Violent Ultimatum",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.B,
				cost.R,
				cost.R,
				cost.R,
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Black, color.Green, color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 3,
						MaxTargets: 3,
						Constraint: "three target permanents",
						Allow:      game.TargetAllowPermanent,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(1),
						},
					},
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(2),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Destroy three target permanents.
		`,
		},
	}
}
