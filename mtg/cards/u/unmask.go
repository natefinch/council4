package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Unmask is the card definition for Unmask.
//
// Type: Sorcery
// Cost: {3}{B}
//
// Oracle text:
//
//	You may exile a black card from your hand rather than pay this spell's mana cost.
//	Target player reveals their hand. You choose a nonland card from it. That player discards that card.
var Unmask = newUnmask

func newUnmask() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Unmask",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Exile a black card",
					AdditionalCosts: []cost.Additional{
						{
							Kind:           cost.AdditionalExile,
							Amount:         1,
							Source:         zone.Hand,
							MatchCardColor: true,
							CardColor:      color.Black,
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ChooseDiscardFromHand{
							Player:      game.TargetPlayerReference(0),
							ExcludeLand: true,
						},
					},
				},
			}.Ability()),
			OracleText: `
			You may exile a black card from your hand rather than pay this spell's mana cost.
			Target player reveals their hand. You choose a nonland card from it. That player discards that card.
		`,
		},
	}
}
