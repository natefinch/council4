package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Sunscour is the card definition for Sunscour.
//
// Type: Sorcery
// Cost: {5}{W}{W}
//
// Oracle text:
//
//	You may exile two white cards from your hand rather than pay this spell's mana cost.
//	Destroy all creatures.
var Sunscour = newSunscour

func newSunscour() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Sunscour",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Exile 2 white cards",
					AdditionalCosts: []cost.Additional{
						{
							Kind:           cost.AdditionalExile,
							Amount:         2,
							Source:         zone.Hand,
							MatchCardColor: true,
							CardColor:      color.White,
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			You may exile two white cards from your hand rather than pay this spell's mana cost.
			Destroy all creatures.
		`,
		},
	}
}
