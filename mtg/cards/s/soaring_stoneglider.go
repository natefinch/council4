package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SoaringStoneglider is the card definition for Soaring Stoneglider.
//
// Type: Creature — Elephant Cleric
// Cost: {2}{W}
//
// Oracle text:
//
//	As an additional cost to cast this spell, exile two cards from your graveyard or pay {1}{W}.
//	Flying, vigilance
var SoaringStoneglider = newSoaringStoneglider

func newSoaringStoneglider() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Soaring Stoneglider",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elephant, types.Cleric},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
			},
			AdditionalCostChoices: []cost.AdditionalChoice{
				cost.AdditionalChoice{
					Options: []cost.AdditionalChoiceOption{
						cost.AdditionalChoiceOption{
							Label: "Exile two cards from your graveyard",
							Costs: []cost.Additional{
								{
									Kind:   cost.AdditionalExile,
									Text:   "exile two cards from your graveyard",
									Amount: 2,
									Source: zone.Graveyard,
								},
							},
						},
						cost.AdditionalChoiceOption{
							Label: "Pay {1}{W}",
							Mana:  cost.Mana{cost.O(1), cost.W},
						},
					},
				},
			},
			OracleText: `
			As an additional cost to cast this spell, exile two cards from your graveyard or pay {1}{W}.
			Flying, vigilance
		`,
		},
	}
}
