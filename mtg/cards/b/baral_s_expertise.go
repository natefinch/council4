package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BaralSExpertise is the card definition for Baral's Expertise.
//
// Type: Sorcery
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	Return up to three target artifacts and/or creatures to their owners' hands.
//	You may cast a spell with mana value 4 or less from your hand without paying its mana cost.
var BaralSExpertise = newBaralSExpertise

func newBaralSExpertise() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Baral's Expertise",
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
						MinTargets: 0,
						MaxTargets: 3,
						Constraint: "up to three target artifacts and/or creatures",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}}),
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
						Primitive: game.CastForFree{
							Player: game.ControllerReference(),
							Selection: game.Selection{ExcludedTypes: []types.Card{types.Land}, ManaValue: opt.Val(compare.Int{
								Op:    compare.LessOrEqual,
								Value: 4,
							})},
							Zone: zone.Hand,
						},
						Optional: true,
					},
				},
			}.Ability()),
			OracleText: `
			Return up to three target artifacts and/or creatures to their owners' hands.
			You may cast a spell with mana value 4 or less from your hand without paying its mana cost.
		`,
		},
	}
}
