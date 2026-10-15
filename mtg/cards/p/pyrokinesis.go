package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Pyrokinesis is the card definition for Pyrokinesis.
//
// Type: Instant
// Cost: {4}{R}{R}
//
// Oracle text:
//
//	You may exile a red card from your hand rather than pay this spell's mana cost.
//	Pyrokinesis deals 4 damage divided as you choose among any number of target creatures.
var Pyrokinesis = newPyrokinesis

func newPyrokinesis() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Pyrokinesis",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Exile a red card",
					AdditionalCosts: []cost.Additional{
						{
							Kind:           cost.AdditionalExile,
							Amount:         1,
							Source:         zone.Hand,
							MatchCardColor: true,
							CardColor:      color.Red,
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 4,
						Constraint: "any number of target creatures",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(4),
							Recipient: game.AnyTargetDamageRecipient(0),
							Divided:   true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			You may exile a red card from your hand rather than pay this spell's mana cost.
			Pyrokinesis deals 4 damage divided as you choose among any number of target creatures.
		`,
		},
	}
}
