package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Pulverize is the card definition for Pulverize.
//
// Type: Sorcery
// Cost: {4}{R}{R}
//
// Oracle text:
//
//	You may sacrifice two Mountains rather than pay this spell's mana cost.
//	Destroy all artifacts.
var Pulverize = newPulverize

func newPulverize() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Pulverize",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Sacrifice two Mountains",
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalSacrifice,
							Text:        "sacrifice two Mountains",
							Amount:      2,
							SubtypesAny: cost.SubtypeSet{types.Mountain},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			You may sacrifice two Mountains rather than pay this spell's mana cost.
			Destroy all artifacts.
		`,
		},
	}
}
