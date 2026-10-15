package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AetherGale is the card definition for Aether Gale.
//
// Type: Sorcery
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	Return six target nonland permanents to their owners' hands.
var AetherGale = newAetherGale

func newAetherGale() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Aether Gale",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 6,
						MaxTargets: 6,
						Constraint: "six target nonland permanents",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(0),
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(1),
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(2),
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(3),
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(4),
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(5),
							Destination: zone.Hand,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Return six target nonland permanents to their owners' hands.
		`,
		},
	}
}
