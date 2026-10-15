package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LumberingSatyr is the card definition for Lumbering Satyr.
//
// Type: Creature — Satyr Beast
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	All creatures have forestwalk. (They can't be blocked as long as defending player controls a Forest.)
var LumberingSatyr = newLumberingSatyr

func newLumberingSatyr() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Lumbering Satyr",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Satyr, types.Beast},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "All creatures have forestwalk. (They can't be blocked as long as defending player controls a Forest.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.StaticAbility{
									Text: "Forestwalk",
									KeywordAbilities: []game.KeywordAbility{
										game.LandwalkKeyword{
											Subtype: types.Forest,
										},
									},
								},
							},
						},
					},
				},
			},
			OracleText: `
			All creatures have forestwalk. (They can't be blocked as long as defending player controls a Forest.)
		`,
		},
	}
}
