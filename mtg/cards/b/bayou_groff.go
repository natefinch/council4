package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BayouGroff is the card definition for Bayou Groff.
//
// Type: Creature — Plant Dog
// Cost: {1}{G}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice a creature or pay {3}.
var BayouGroff = newBayouGroff

func newBayouGroff() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Bayou Groff",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Plant, types.Dog},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 4}),
			AdditionalCostChoices: []cost.AdditionalChoice{
				cost.AdditionalChoice{
					Options: []cost.AdditionalChoiceOption{
						cost.AdditionalChoiceOption{
							Label: "Sacrifice a creature",
							Costs: []cost.Additional{
								{
									Kind:               cost.AdditionalSacrifice,
									Text:               "sacrifice a creature",
									Amount:             1,
									MatchPermanentType: true,
									PermanentType:      types.Creature,
								},
							},
						},
						cost.AdditionalChoiceOption{
							Label: "Pay {3}",
							Mana:  cost.Mana{cost.O(3)},
						},
					},
				},
			},
			OracleText: `
			As an additional cost to cast this spell, sacrifice a creature or pay {3}.
		`,
		},
	}
}
