package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Delraich is the card definition for Delraich.
//
// Type: Creature — Horror
// Cost: {6}{B}
//
// Oracle text:
//
//	You may sacrifice three black creatures rather than pay this spell's mana cost.
//	Trample
var Delraich = newDelraich

func newDelraich() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Delraich",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Horror},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Sacrifice three black creatures",
					AdditionalCosts: []cost.Additional{
						{
							Kind:               cost.AdditionalSacrifice,
							Text:               "sacrifice three black creatures",
							Amount:             3,
							MatchPermanentType: true,
							PermanentType:      types.Creature,
							MatchCardColor:     true,
							CardColor:          color.Black,
						},
					},
				},
			},
			OracleText: `
			You may sacrifice three black creatures rather than pay this spell's mana cost.
			Trample
		`,
		},
	}
}
