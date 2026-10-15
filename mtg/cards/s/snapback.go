package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Snapback is the card definition for Snapback.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	You may exile a blue card from your hand rather than pay this spell's mana cost.
//	Return target creature to its owner's hand.
var Snapback = newSnapback

func newSnapback() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Snapback",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Exile a blue card",
					AdditionalCosts: []cost.Additional{
						{
							Kind:           cost.AdditionalExile,
							Amount:         1,
							Source:         zone.Hand,
							MatchCardColor: true,
							CardColor:      color.Blue,
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(0),
							Destination: zone.Hand,
						},
					},
				},
			}.Ability()),
			OracleText: `
			You may exile a blue card from your hand rather than pay this spell's mana cost.
			Return target creature to its owner's hand.
		`,
		},
	}
}
