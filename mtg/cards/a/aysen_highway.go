package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AysenHighway is the card definition for Aysen Highway.
//
// Type: Enchantment
// Cost: {3}{W}{W}{W}
//
// Oracle text:
//
//	White creatures have plainswalk. (They can't be blocked as long as defending player controls a Plains.)
var AysenHighway = newAysenHighway

func newAysenHighway() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Aysen Highway",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "White creatures have plainswalk. (They can't be blocked as long as defending player controls a Plains.)",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.White}}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.StaticAbility{
									Text: "Plainswalk",
									KeywordAbilities: []game.KeywordAbility{
										game.LandwalkKeyword{
											Subtype: types.Plains,
										},
									},
								},
							},
						},
					},
				},
			},
			OracleText: `
			White creatures have plainswalk. (They can't be blocked as long as defending player controls a Plains.)
		`,
		},
	}
}
