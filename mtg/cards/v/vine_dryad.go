package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// VineDryad is the card definition for Vine Dryad.
//
// Type: Creature — Dryad
// Cost: {3}{G}
//
// Oracle text:
//
//	You may exile a green card from your hand rather than pay this spell's mana cost.
//	Flash
//	Forestwalk (This creature can't be blocked as long as defending player controls a Forest.)
var VineDryad = newVineDryad

func newVineDryad() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Vine Dryad",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dryad},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.ForestwalkStaticBody,
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Exile a green card",
					AdditionalCosts: []cost.Additional{
						{
							Kind:           cost.AdditionalExile,
							Amount:         1,
							Source:         zone.Hand,
							MatchCardColor: true,
							CardColor:      color.Green,
						},
					},
				},
			},
			OracleText: `
			You may exile a green card from your hand rather than pay this spell's mana cost.
			Flash
			Forestwalk (This creature can't be blocked as long as defending player controls a Forest.)
		`,
		},
	}
}
